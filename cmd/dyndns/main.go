package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// DynDNSRequest represents the request body to IONOS
type DynDNSRequest struct {
	Domains     []string `json:"domains"`
	Description string   `json:"description"`
}

const defaultAPIURL = "https://api.hosting.ionos.com/dns/v1/dyndns"

const (
	// requestTimeout bounds a single DNS update call so a stalled API cannot
	// block the update loop forever.
	requestTimeout = 30 * time.Second
	// maxResponseBytes caps how much of the API response is read into memory.
	maxResponseBytes = 1 << 20 // 1 MiB
	// maxLoggedBodyBytes caps how much of the response body reaches the logs.
	maxLoggedBodyBytes = 512
)

// httpClient is shared across updates so connections are reused and every
// request inherits the same timeout and TLS floor.
var httpClient = &http.Client{
	Timeout: requestTimeout,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	},
}

// truncateBody shortens a response body for safe logging.
func truncateBody(body []byte) string {
	if len(body) > maxLoggedBodyBytes {
		return string(body[:maxLoggedBodyBytes]) + "...(truncated)"
	}
	return string(body)
}

// DynDNSResponse is returned by the IONOS DynDNS registration endpoint.
// Registering does not change any record: the IP is only updated when
// UpdateURL is called, IONOS using the caller's source address.
type DynDNSResponse struct {
	BulkID    string `json:"bulkId"`
	UpdateURL string `json:"updateUrl"`
}

// apiStatusError reports a non-200 answer from the IONOS API.
type apiStatusError struct {
	status int
}

func (e *apiStatusError) Error() string {
	return fmt.Sprintf("IONOS API error: status %d", e.status)
}

// registerDynDNS registers the domains with IONOS and returns the update URL
// to call to push the current public IP.
func registerDynDNS(ctx context.Context, config Config, apiURL string) (string, error) {
	// Build the request body
	reqBody := DynDNSRequest{
		Domains:     config.Domains,
		Description: "IONOS DynDNS Updater",
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	slog.Debug("Sending DynDNS registration request", "domains", config.Domains)

	// Create HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		apiURL,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", err
	}

	// Add headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.APIKey)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read the response, bounded to avoid unbounded allocation
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", err
	}

	// A successful body holds the secret update URL, so only its status is logged.
	slog.Debug("Registration response received", "status", resp.StatusCode)

	// Check the status
	if resp.StatusCode != http.StatusOK {
		slog.Debug("Registration error body", "body", truncateBody(body))
		return "", &apiStatusError{status: resp.StatusCode}
	}

	var parsed DynDNSResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("invalid registration response: %w", err)
	}
	if parsed.UpdateURL == "" {
		return "", errors.New("registration response has no updateUrl")
	}

	slog.Debug("DynDNS registered", "bulk_id", parsed.BulkID)

	return parsed.UpdateURL, nil
}

// triggerUpdate calls the update URL so IONOS sets the records to the public
// IP the request comes from. The URL embeds a secret token and must never end
// up in logs or error messages.
func triggerUpdate(ctx context.Context, updateURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateURL, nil)
	if err != nil {
		return errors.New("invalid update URL")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		// *url.Error embeds the full URL, token included: keep only the cause.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return fmt.Errorf("update request failed: %w", urlErr.Err)
		}
		return errors.New("update request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return err
	}

	slog.Debug("Update response received", "status", resp.StatusCode, "body", truncateBody(body))

	if resp.StatusCode != http.StatusOK {
		return &apiStatusError{status: resp.StatusCode}
	}

	return nil
}

// isStaleUpdateURL reports whether an update failure means the update URL is
// no longer valid and a new registration is needed.
func isStaleUpdateURL(err error) bool {
	var statusErr *apiStatusError
	if !errors.As(err, &statusErr) {
		return false
	}
	switch statusErr.status {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return true
	}
	return false
}

// healthMux builds the handler set exposed by the health check server.
func healthMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "ok")
	})
	return mux
}

// newHealthServer returns a health check server with timeouts set on every
// stage of the request lifecycle (gosec G114 / Slowloris).
func newHealthServer(port int) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           healthMux(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// shutdownTimeout bounds how long the health server gets to drain on exit.
const shutdownTimeout = 5 * time.Second

// dnsUpdater keeps the state shared between successive updates.
type dnsUpdater struct {
	config Config
	apiURL string
	// updateURL is reused across ticks so a new IONOS bulk is not created on
	// every interval.
	updateURL string
	// lastOK makes successes log at INFO only after startup or a failure, so
	// the logs show the service works without one line per interval.
	lastOK bool
}

// update registers with IONOS if needed, then pushes the current public IP.
func (u *dnsUpdater) update(ctx context.Context) error {
	if u.updateURL == "" {
		regCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		updateURL, err := registerDynDNS(regCtx, u.config, u.apiURL)
		cancel()
		if err != nil {
			return fmt.Errorf("registration: %w", err)
		}
		u.updateURL = updateURL
	}

	updCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	if err := triggerUpdate(updCtx, u.updateURL); err != nil {
		if isStaleUpdateURL(err) {
			u.updateURL = ""
		}
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

// updateAndLog runs an update and logs its outcome.
func (u *dnsUpdater) updateAndLog(ctx context.Context) error {
	err := u.update(ctx)
	if err != nil {
		slog.Error("DNS update failed", "error", err)
		u.lastOK = false
		return err
	}
	if !u.lastOK {
		slog.Info("DNS updated with current public IP", "domains", u.config.Domains)
	} else {
		slog.Debug("DNS updated with current public IP", "domains", u.config.Domains)
	}
	u.lastOK = true
	return nil
}

// runLoop performs the initial update then repeats it on every tick until the
// context is cancelled.
func runLoop(ctx context.Context, config Config, apiURL string) {
	updater := &dnsUpdater{config: config, apiURL: apiURL}

	// First immediate update
	_ = updater.updateAndLog(ctx)

	// Heartbeat
	heartbeatInterval := time.Duration(config.HeartbeatInterval) * time.Second
	lastHeartbeat := time.Now()
	updateCount := 0

	// Periodic loop
	ticker := time.NewTicker(time.Duration(config.UpdateInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutdown signal received, stopping update loop")
			return
		case <-ticker.C:
			if err := updater.updateAndLog(ctx); err == nil {
				updateCount++
			}

			if time.Since(lastHeartbeat) >= heartbeatInterval {
				slog.Info("Heartbeat: service running", "successful_updates_since_last", updateCount)
				lastHeartbeat = time.Now()
				updateCount = 0
			}
		}
	}
}

// run holds the application logic and reports failures to main, which owns the
// process exit code.
func run() error {
	logLevel := setupLogger()
	slog.Info("IONOS DynDNS starting", "log_level", logLevel.String())

	config := loadConfig()

	if config.APIKey == "" {
		return errors.New("IONOS_API_KEY not defined")
	}
	if len(config.Domains) == 0 {
		return errors.New("IONOS_DOMAINS not defined")
	}

	slog.Info("Configuration loaded",
		"domains", config.Domains,
		"update_interval_seconds", config.UpdateInterval,
		"heartbeat_interval_seconds", config.HeartbeatInterval,
		"health_port", config.HealthPort,
	)

	// Cancelled on SIGINT/SIGTERM so in-flight requests are aborted on exit.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start health check server on a dedicated mux so that no package can
	// register extra handlers (e.g. net/http/pprof) on this listener.
	srv := newHealthServer(config.HealthPort)
	go func() {
		slog.Info("Health check server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Health check server failed", "error", err)
		}
	}()

	runLoop(ctx, config, defaultAPIURL)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Health check server shutdown failed", "error", err)
	}

	slog.Info("Shutdown complete")
	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("Fatal error", "error", err)
		os.Exit(1)
	}
}
