package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDynDNSRequestJSON(t *testing.T) {
	req := DynDNSRequest{
		Domains:     []string{"example.com", "sub.example.com"},
		Description: "IONOS DynDNS Updater",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	domains, ok := decoded["domains"].([]interface{})
	if !ok {
		t.Fatal("expected domains to be a list")
	}
	if len(domains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(domains))
	}
	if domains[0] != "example.com" {
		t.Errorf("expected first domain to be example.com, got %s", domains[0])
	}
}

func TestUpdateDNSSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected API key header, got %s", r.Header.Get("X-API-Key"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected JSON content type, got %s", r.Header.Get("Content-Type"))
		}

		var body DynDNSRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if len(body.Domains) != 1 || body.Domains[0] != "example.com" {
			t.Errorf("unexpected domains: %v", body.Domains)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Override the API URL by using the test server
	config := Config{
		APIKey:  "test-key",
		Domains: []string{"example.com"},
	}

	err := updateDNSWithURL(config, server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateDNSAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	config := Config{
		APIKey:  "bad-key",
		Domains: []string{"example.com"},
	}

	err := updateDNSWithURL(config, server.URL)
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
}
