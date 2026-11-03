package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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
		_, _ = fmt.Fprint(w, `{"bulkId":"b1","updateUrl":"https://ipv4.example.test/dyndns?q=secret"}`)
	}))
	defer server.Close()

	// Override the API URL by using the test server
	config := Config{
		APIKey:  "test-key",
		Domains: []string{"example.com"},
	}

	updateURL, err := registerDynDNS(t.Context(), config, server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updateURL != "https://ipv4.example.test/dyndns?q=secret" {
		t.Errorf("unexpected update URL: %q", updateURL)
	}
}

func TestRegisterDynDNSMissingUpdateURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"bulkId":"b1"}`)
	}))
	defer server.Close()

	config := Config{APIKey: "test-key", Domains: []string{"example.com"}}

	if _, err := registerDynDNS(t.Context(), config, server.URL); err == nil {
		t.Fatal("expected error when updateUrl is missing, got nil")
	}
}

// fakeIONOS serves the registration endpoint on / and the update URL on
// /update, counting calls to each and answering updates with updateStatus.
type fakeIONOS struct {
	server        *httptest.Server
	registrations atomic.Int32
	updates       atomic.Int32
	updateStatus  atomic.Int32
}

func newFakeIONOS(t *testing.T) *fakeIONOS {
	t.Helper()
	f := &fakeIONOS{}
	f.updateStatus.Store(http.StatusOK)
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/":
			f.registrations.Add(1)
			_, _ = fmt.Fprintf(w, `{"bulkId":"b1","updateUrl":"%s/update?q=secret"}`, f.server.URL)
		case r.Method == http.MethodGet && r.URL.Path == "/update":
			if r.URL.Query().Get("q") != "secret" {
				t.Errorf("update URL token not forwarded, got %q", r.URL.RawQuery)
			}
			f.updates.Add(1)
			w.WriteHeader(int(f.updateStatus.Load()))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(f.server.Close)
	return f
}

func TestUpdaterRegistersOnceAndTriggersEveryTime(t *testing.T) {
	f := newFakeIONOS(t)
	u := &dnsUpdater{
		config: Config{APIKey: "test-key", Domains: []string{"example.com"}},
		apiURL: f.server.URL,
	}

	for i := range 3 {
		if err := u.update(t.Context()); err != nil {
			t.Fatalf("update %d failed: %v", i, err)
		}
	}

	if got := f.registrations.Load(); got != 1 {
		t.Errorf("expected 1 registration, got %d", got)
	}
	if got := f.updates.Load(); got != 3 {
		t.Errorf("expected 3 update calls, got %d", got)
	}
}

func TestUpdaterReportsUpdateFailure(t *testing.T) {
	f := newFakeIONOS(t)
	f.updateStatus.Store(http.StatusInternalServerError)
	u := &dnsUpdater{
		config: Config{APIKey: "test-key", Domains: []string{"example.com"}},
		apiURL: f.server.URL,
	}

	if err := u.update(t.Context()); err == nil {
		t.Fatal("expected error for 500 update response, got nil")
	}
	// A server error is transient: the update URL is kept.
	if u.updateURL == "" {
		t.Error("update URL should be kept after a transient failure")
	}
}

func TestUpdaterReregistersAfterRejectedUpdateURL(t *testing.T) {
	f := newFakeIONOS(t)
	f.updateStatus.Store(http.StatusUnauthorized)
	u := &dnsUpdater{
		config: Config{APIKey: "test-key", Domains: []string{"example.com"}},
		apiURL: f.server.URL,
	}

	if err := u.update(t.Context()); err == nil {
		t.Fatal("expected error for 401 update response, got nil")
	}

	f.updateStatus.Store(http.StatusOK)
	if err := u.update(t.Context()); err != nil {
		t.Fatalf("expected recovery after re-registration, got %v", err)
	}
	if got := f.registrations.Load(); got != 2 {
		t.Errorf("expected 2 registrations, got %d", got)
	}
}

func TestTriggerUpdateErrorHidesToken(t *testing.T) {
	// Nothing listens on this port, so the request fails at connection time.
	err := triggerUpdate(t.Context(), "http://127.0.0.1:1/update?q=supersecret")
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
	if strings.Contains(err.Error(), "supersecret") {
		t.Errorf("error leaks the update token: %v", err)
	}
}

func TestUpdateDNSAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"error":"unauthorized"}`)
	}))
	defer server.Close()

	config := Config{
		APIKey:  "bad-key",
		Domains: []string{"example.com"},
	}

	_, err := registerDynDNS(t.Context(), config, server.URL)
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
}

func TestHealthEndpoint(t *testing.T) {
	// Exercise the real handler set, not a copy of it.
	server := httptest.NewServer(healthMux())
	defer server.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/healthz", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to call /healthz: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != "ok" {
		t.Errorf("expected 'ok', got '%s'", string(body))
	}
}

func TestHealthServerHasTimeouts(t *testing.T) {
	srv := newHealthServer(8080)

	if srv.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %s", srv.Addr)
	}
	if srv.ReadHeaderTimeout == 0 {
		t.Error("ReadHeaderTimeout must be set to mitigate Slowloris")
	}
	if srv.ReadTimeout == 0 || srv.WriteTimeout == 0 || srv.IdleTimeout == 0 {
		t.Error("read, write and idle timeouts must all be set")
	}
	if srv.Handler == nil {
		t.Error("server must use a dedicated handler, not DefaultServeMux")
	}
}

func TestHealthMuxRejectsUnknownPaths(t *testing.T) {
	server := httptest.NewServer(healthMux())
	defer server.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/debug/pprof/", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for unregistered path, got %d", resp.StatusCode)
	}
}

func TestUpdateDNSRespectsContextCancellation(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	defer close(release)

	config := Config{APIKey: "test-key", Domains: []string{"example.com"}}

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := registerDynDNS(ctx, config, server.URL)
	if err == nil {
		t.Fatal("expected error when context deadline is exceeded, got nil")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("request was not cancelled promptly, took %v", elapsed)
	}
}

func TestHTTPClientHasTimeout(t *testing.T) {
	if httpClient.Timeout == 0 {
		t.Error("shared HTTP client must define a timeout")
	}
}

func TestTruncateBody(t *testing.T) {
	short := []byte("ok")
	if got := truncateBody(short); got != "ok" {
		t.Errorf("expected short body unchanged, got %q", got)
	}

	long := bytes.Repeat([]byte("a"), maxLoggedBodyBytes*2)
	got := truncateBody(long)
	if len(got) >= len(long) {
		t.Errorf("expected long body to be truncated, got %d bytes", len(got))
	}
	if !strings.HasSuffix(got, "...(truncated)") {
		t.Error("expected truncation marker in output")
	}
}

func TestUpdateDNSBoundsResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bytes.Repeat([]byte("a"), maxResponseBytes+4096))
	}))
	defer server.Close()

	if err := triggerUpdate(t.Context(), server.URL); err != nil {
		t.Fatalf("expected oversized body to be tolerated, got %v", err)
	}
}

func TestRunLoopStopsOnContextCancellation(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := Config{
		APIKey:            "test-key",
		Domains:           []string{"example.com"},
		UpdateInterval:    3600,
		HeartbeatInterval: 3600,
	}

	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan struct{})
	go func() {
		defer close(done)
		runLoop(ctx, config, server.URL)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("runLoop did not return after context cancellation")
	}
}
