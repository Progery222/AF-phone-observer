package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunDeletesCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/cache/stub" {
			t.Fatalf("expected /cache/stub, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("priority") != "high" || r.URL.Query().Get("timeout_sec") != "5" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"serial":"stub","cleared":true,"screen_cleared":true,"ui_cleared":true,"cleared_at":"2026-06-23T09:00:00Z"}`))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), []string{"-url=" + server.URL, "-serial=stub", "-priority=high", "-timeout-sec=5", "-auto-start=false"}, func(string) string { return "" }, &stdout, &stderr, server.Client())
	if err != nil {
		t.Fatalf("run failed: %v, stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"cleared": true`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestRunRequiresSerial(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), nil, func(string) string { return "" }, &stdout, &stderr, http.DefaultClient)
	if err == nil || !strings.Contains(err.Error(), "SERIAL обязателен") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCacheEndpointDoesNotDuplicatePath(t *testing.T) {
	got := cacheEndpoint("http://localhost:9090/cache/stub", "stub")
	if got != "http://localhost:9090/cache/stub" {
		t.Fatalf("unexpected endpoint: %s", got)
	}
}
