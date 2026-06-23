package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunGetsScreen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/screen/stub" {
			t.Fatalf("expected /screen/stub, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("priority") != "high" || r.URL.Query().Get("timeout_sec") != "12" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"serial":"stub","screenshot_url":"af-screenshots/stub/shot.png","minio_key":"stub/shot.png","size_bytes":67,"resolution":{"width":1,"height":1},"taken_at":"2026-06-23T09:00:00Z"}`))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), []string{"-url=" + server.URL, "-serial=stub", "-priority=high", "-timeout-sec=12", "-auto-start=false"}, func(string) string { return "" }, &stdout, &stderr, server.Client())
	if err != nil {
		t.Fatalf("run failed: %v, stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"screenshot_url": "af-screenshots/stub/shot.png"`) {
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

func TestScreenEndpointDoesNotDuplicatePath(t *testing.T) {
	got := screenEndpoint("http://localhost:9090/screen/stub", "stub")
	if got != "http://localhost:9090/screen/stub" {
		t.Fatalf("unexpected endpoint: %s", got)
	}
}
