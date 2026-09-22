package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

func updateDNS(ctx context.Context, config Config) error {
	return updateDNSWithURL(ctx, config, defaultAPIURL)
}

func updateDNSWithURL(ctx context.Context, config Config, apiURL string) error {
	// Build the request body
	reqBody := DynDNSRequest{
		Domains:     config.Domains,
		Description: "IONOS DynDNS Updater",
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	slog.Debug("Sending DNS update request", "domains", config.Domains)

	// Create HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		apiURL,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return err
	}

	// Add headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.APIKey)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read and display the response, bounded to avoid unbounded allocation
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return err
	}

	slog.Debug("API response received", "status", resp.StatusCode, "body", truncateBody(body))

	// Check the status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IONOS API error: status %d", resp.StatusCode)
	}

	slog.Debug("DNS updated successfully", "domains", config.Domains, "status", resp.StatusCode)

	return nil
}

func main() {
	logLevel := setupLogger()
	slog.Info("IONOS DynDNS starting", "log_level", logLevel.String())

	config := loadConfig()

	if config.APIKey == "" {
		slog.Error("IONOS_API_KEY not defined")
		return
	}
	if len(config.Domains) == 0 || config.Domains[0] == "" {
		slog.Error("IONOS_DOMAINS not defined")
		return
	}

	slog.Info("Configuration loaded",
		"domains", config.Domains,
		"update_interval_seconds", config.UpdateInterval,
		"heartbeat_interval_seconds", config.HeartbeatInterval,
		"health_port", config.HealthPort,
	)

	// Start health check server
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "ok")
	})
	go func() {
		addr := fmt.Sprintf(":%d", config.HealthPort)
		slog.Info("Health check server starting", "addr", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			slog.Error("Health check server failed", "error", err)
		}
	}()

	// First immediate update
	if err := updateDNS(context.Background(), config); err != nil {
		slog.Error("DNS update failed", "error", err)
	}

	// Heartbeat
	heartbeatInterval := time.Duration(config.HeartbeatInterval) * time.Second
	lastHeartbeat := time.Now()
	updateCount := 0

	// Periodic loop
	ticker := time.NewTicker(time.Duration(config.UpdateInterval) * time.Second)
	for range ticker.C {
		if err := updateDNS(context.Background(), config); err != nil {
			slog.Error("DNS update failed", "error", err)
		} else {
			updateCount++
		}

		if time.Since(lastHeartbeat) >= heartbeatInterval {
			slog.Info("Heartbeat: service running", "successful_updates_since_last", updateCount)
			lastHeartbeat = time.Now()
			updateCount = 0
		}
	}
}
