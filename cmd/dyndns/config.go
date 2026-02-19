package main

// Config contains the application configuration
type Config struct {
	APIKey            string   // IONOS API key (from environment)
	Domains           []string // List of domains to update
	UpdateInterval    int      // Interval in seconds between each update
	HeartbeatInterval int      // Interval in seconds between heartbeat logs
}
