package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// DynDNSRequest represents the request body to IONOS
type DynDNSRequest struct {
	Domains     []string `json:"domains"`
	Description string   `json:"description"`
}

const defaultAPIURL = "https://api.hosting.ionos.com/dns/v1/dyndns"

func updateDNS(config Config) error {
	return updateDNSWithURL(config, defaultAPIURL)
}

func updateDNSWithURL(config Config, apiURL string) error {
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
	req, err := http.NewRequest(
		"POST",
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
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read and display the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	slog.Debug("API response received", "status", resp.StatusCode, "body", string(body))

	// Check the status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IONOS API error: status %d", resp.StatusCode)
	}

	slog.Debug("DNS updated successfully", "domains", config.Domains, "status", resp.StatusCode)

	return nil
}

func main() {
	// Configure slog based on LOG_LEVEL environment variable
	logLevel := slog.LevelInfo
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		}
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	slog.SetDefault(slog.New(handler))

	slog.Info("IONOS DynDNS starting", "log_level", logLevel.String())

	// Read interval (default: 300 seconds = 5 minutes)
	interval := 300
	if envInterval := os.Getenv("UPDATE_INTERVAL_SECONDS"); envInterval != "" {
		if parsed, err := strconv.Atoi(envInterval); err == nil {
			interval = parsed
		}
	}

	// Read heartbeat interval (default: 21600 seconds = 6 hours)
	heartbeatSecs := 21600
	if envHeartbeat := os.Getenv("HEARTBEAT_INTERVAL_SECONDS"); envHeartbeat != "" {
		if parsed, err := strconv.Atoi(envHeartbeat); err == nil {
			heartbeatSecs = parsed
		}
	}

	// Read health port (default: 8080)
	healthPort := 8080
	if envPort := os.Getenv("HEALTH_PORT"); envPort != "" {
		if parsed, err := strconv.Atoi(envPort); err == nil {
			healthPort = parsed
		}
	}

	// Load config from environment
	config := Config{
		APIKey:            os.Getenv("IONOS_API_KEY"),
		Domains:           strings.Split(os.Getenv("IONOS_DOMAINS"), ","),
		UpdateInterval:    interval,
		HeartbeatInterval: heartbeatSecs,
		HealthPort:        healthPort,
	}

	// Check that API key is present
	if config.APIKey == "" {
		slog.Error("IONOS_API_KEY not defined")
		return
	}

	// Check that we have at least one domain
	if len(config.Domains) == 0 || config.Domains[0] == "" {
		slog.Error("IONOS_DOMAINS not defined")
		return
	}

	slog.Info("Configuration loaded",
		"domains", config.Domains,
		"update_interval_seconds", interval,
		"heartbeat_interval_seconds", heartbeatSecs,
		"health_port", healthPort,
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
	if err := updateDNS(config); err != nil {
		slog.Error("DNS update failed", "error", err)
	}

	// Heartbeat
	heartbeatInterval := time.Duration(config.HeartbeatInterval) * time.Second
	lastHeartbeat := time.Now()
	updateCount := 0

	// Periodic loop
	ticker := time.NewTicker(time.Duration(config.UpdateInterval) * time.Second)
	for range ticker.C {
		if err := updateDNS(config); err != nil {
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
