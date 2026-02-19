package main

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config contains the application configuration
type Config struct {
	APIKey            string   // IONOS API key (from environment)
	Domains           []string // List of domains to update
	UpdateInterval    int      // Interval in seconds between each update
	HeartbeatInterval int      // Interval in seconds between heartbeat logs
	HealthPort        int      // Port for the health check endpoint
}

func envInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func setupLogger() slog.Level {
	logLevel := slog.LevelInfo
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			logLevel = slog.LevelDebug
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

	return logLevel
}

func loadConfig() Config {
	return Config{
		APIKey:            os.Getenv("IONOS_API_KEY"),
		Domains:           strings.Split(os.Getenv("IONOS_DOMAINS"), ","),
		UpdateInterval:    envInt("UPDATE_INTERVAL_SECONDS", 300),
		HeartbeatInterval: envInt("HEARTBEAT_INTERVAL_SECONDS", 21600),
		HealthPort:        envInt("HEALTH_PORT", 8080),
	}
}
