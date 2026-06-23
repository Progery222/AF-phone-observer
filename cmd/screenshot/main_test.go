package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunPostsScreenshotRequest(t *testing.T) {
	var got screenshotRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/screenshot" {
			t.Fatalf("expected /screenshot, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"screenshot_url":"af-screenshots/stub/20260623-120000.png",
			"minio_key":"stub/20260623-120000.png",
			"size_bytes":70,
			"resolution":{"width":1,"height":1},
			"taken_at":"2026-06-23T09:00:00Z"
		}`))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(
		context.Background(),
		[]string{"-url=" + server.URL, "-serial=stub", "-priority=high", "-timeout-sec=3"},
		func(string) string { return "" },
		&stdout,
		&stderr,
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run failed: %v, stderr=%s", err, stderr.String())
	}
	if got.Serial != "stub" || got.Priority != "high" || got.TimeoutSec != 3 || !got.StoreInMinio {
		t.Fatalf("unexpected request payload: %+v", got)
	}
	if !strings.Contains(stdout.String(), `"screenshot_url": "af-screenshots/stub/20260623-120000.png"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestRunRequiresSerial(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), nil, func(string) string { return "" }, &stdout, &stderr, http.DefaultClient)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "SERIAL обязателен") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScreenshotEndpointDoesNotDuplicatePath(t *testing.T) {
	got := screenshotEndpoint("http://localhost:9090/screenshot")
	if got != "http://localhost:9090/screenshot" {
		t.Fatalf("unexpected endpoint: %s", got)
	}
}
