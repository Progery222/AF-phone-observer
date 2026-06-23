package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunGetsUI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/ui/stub" {
			t.Fatalf("expected /ui/stub, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("format") != "xml" || r.URL.Query().Get("priority") != "high" || r.URL.Query().Get("timeout_sec") != "30" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"serial":"stub","xml_dump":"<hierarchy/>","element_count":1,"package_name":"com.example","taken_at":"2026-06-23T09:00:00Z"}`))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), []string{"-url=" + server.URL, "-serial=stub", "-format=xml", "-priority=high", "-timeout-sec=30", "-auto-start=false"}, func(string) string { return "" }, &stdout, &stderr, server.Client())
	if err != nil {
		t.Fatalf("run failed: %v, stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"xml_dump": "\u003chierarchy/\u003e"`) && !strings.Contains(stdout.String(), `"xml_dump": "<hierarchy/>"`) {
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

func TestUIEndpointDoesNotDuplicatePath(t *testing.T) {
	got := uiEndpoint("http://localhost:9090/ui/stub", "stub")
	if got != "http://localhost:9090/ui/stub" {
		t.Fatalf("unexpected endpoint: %s", got)
	}
}
