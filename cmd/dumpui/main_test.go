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

func TestRunPostsDumpUIRequest(t *testing.T) {
	var got dumpUIRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/dump-ui" {
			t.Fatalf("expected /dump-ui, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"elements":[{"type":"Button","text":"OK","resource_id":"id/ok","hint":"","bounds":{"x1":0,"y1":0,"x2":2,"y2":2},"center":{"x":1,"y":1}}],
			"element_count":1,
			"taken_at":"2026-06-23T09:00:00Z"
		}`))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(
		context.Background(),
		[]string{"-url=" + server.URL, "-serial=stub", "-format=json", "-priority=high", "-timeout-sec=30"},
		func(string) string { return "" },
		&stdout,
		&stderr,
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run failed: %v, stderr=%s", err, stderr.String())
	}
	if got.Serial != "stub" || got.Format != "json" || got.Priority != "high" || got.TimeoutSec != 30 {
		t.Fatalf("unexpected request payload: %+v", got)
	}
	if !strings.Contains(stdout.String(), `"element_count": 1`) {
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

func TestDumpUIEndpointDoesNotDuplicatePath(t *testing.T) {
	got := dumpUIEndpoint("http://localhost:9090/dump-ui")
	if got != "http://localhost:9090/dump-ui" {
		t.Fatalf("unexpected endpoint: %s", got)
	}
}
