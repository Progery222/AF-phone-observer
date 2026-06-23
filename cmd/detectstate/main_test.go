package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetectStateCLIPostsRequest(t *testing.T) {
	var got detectStateRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/detect-state" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"state":"unknown","confidence":0,"source":"ui","backend_used":"ui","description":"ok","elements":[],"matched_signals":[],"flags":{"captcha":false,"ban":false,"error":false},"suggested_action":"","package_name":"","element_count":0,"screenshot_url":"","minio_key":"","taken_at":"2026-06-23T09:00:00Z"}`))
	}))
	defer server.Close()

	var stdout strings.Builder
	err := run(context.Background(), []string{
		"-url", server.URL,
		"-serial", "stub",
		"-mode", "ui",
		"-platform", "instagram",
		"-use-screenshot=false",
		"-store-screenshot=true",
		"-priority", "high",
		"-timeout-sec", "12",
		"-auto-start=false",
	}, func(string) string { return "" }, &stdout, &strings.Builder{}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if got.Serial != "stub" || got.Mode != "ui" || got.Platform != "instagram" || got.UseScreenshot || !got.StoreScreenshot || got.Priority != "high" || got.TimeoutSec != 12 {
		t.Fatalf("unexpected request: %+v", got)
	}
	if !strings.Contains(stdout.String(), `"state": "unknown"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestDetectStateCLIValidation(t *testing.T) {
	err := run(context.Background(), []string{"-serial", "", "-auto-start=false"}, func(string) string { return "" }, &strings.Builder{}, &strings.Builder{}, http.DefaultClient)
	if err == nil {
		t.Fatal("expected missing serial error")
	}
}
