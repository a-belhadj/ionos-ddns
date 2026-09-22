package main

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// maxIntervalSeconds caps configurable intervals at roughly one year.
const maxIntervalSeconds = 31536000

// Config contains the application configuration
type Config struct {
	APIKey            string   // IONOS API key (from environment)
	Domains           []string // List of domains to update
	UpdateInterval    int      // Interval in seconds between each update
	HeartbeatInterval int      // Interval in seconds between heartbeat logs
	HealthPort        int      // Port for the health check endpoint
}

// envInt reads an integer environment variable, falling back to defaultVal
// when unset, unparseable or outside [minVal, maxVal].
func envInt(key string, defaultVal, minVal, maxVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		slog.Warn("Invalid integer value, using default", "key", key, "value", val, "default", defaultVal)
		return defaultVal
	}
	if parsed < minVal || parsed > maxVal {
		slog.Warn("Value out of range, using default",
			"key", key, "value", parsed, "min", minVal, "max", maxVal, "default", defaultVal)
		return defaultVal
	}

	return parsed
}

// parseDomains splits and cleans the comma-separated domain list, dropping
// empty entries left by trailing commas or stray whitespace.
func parseDomains(raw string) []string {
	parts := strings.Split(raw, ",")
	domains := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			domains = append(domains, trimmed)
		}
	}
	return domains
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
		APIKey:            strings.TrimSpace(os.Getenv("IONOS_API_KEY")),
		Domains:           parseDomains(os.Getenv("IONOS_DOMAINS")),
		UpdateInterval:    envInt("UPDATE_INTERVAL_SECONDS", 300, 1, maxIntervalSeconds),
		HeartbeatInterval: envInt("HEARTBEAT_INTERVAL_SECONDS", 21600, 1, maxIntervalSeconds),
		HealthPort:        envInt("HEALTH_PORT", 8080, 1, 65535),
	}
}
