package main

import (
	"testing"
)

func TestParseDomains(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{"simple", "example.com,sub.example.com", []string{"example.com", "sub.example.com"}},
		{"spaces", " example.com , sub.example.com ", []string{"example.com", "sub.example.com"}},
		{"trailing comma", "example.com,", []string{"example.com"}},
		{"empty entries", "example.com,,sub.example.com", []string{"example.com", "sub.example.com"}},
		{"empty string", "", nil},
		{"only separators", " , , ", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDomains(tt.raw)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("index %d: expected %q, got %q", i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestEnvIntBounds(t *testing.T) {
	const key = "TEST_ENV_INT"

	tests := []struct {
		name  string
		value string
		set   bool
		want  int
	}{
		{"unset falls back", "", false, 8080},
		{"empty falls back", "", true, 8080},
		{"valid value", "9090", true, 9090},
		{"not a number", "abc", true, 8080},
		{"below range", "0", true, 8080},
		{"negative", "-1", true, 8080},
		{"above range", "65536", true, 8080},
		{"upper bound accepted", "65535", true, 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv(key, tt.value)
			}
			if got := envInt(key, 8080, 1, 65535); got != tt.want {
				t.Errorf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("IONOS_API_KEY", "  secret-key  ")
	t.Setenv("IONOS_DOMAINS", "example.com, sub.example.com,")

	config := loadConfig()

	if config.APIKey != "secret-key" {
		t.Errorf("expected trimmed API key, got %q", config.APIKey)
	}
	if len(config.Domains) != 2 {
		t.Fatalf("expected 2 domains, got %v", config.Domains)
	}
	if config.UpdateInterval != 300 || config.HeartbeatInterval != 21600 || config.HealthPort != 8080 {
		t.Errorf("unexpected defaults: %+v", config)
	}
}
