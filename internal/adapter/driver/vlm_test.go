package driver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/config"
)

func TestCascadingScreenAnalyzerUsesOllamaBackend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		writeTestJSON(t, w, map[string]any{
			"response": `{"screen_type":"login","confidence":0.82,"description":"Login screen","elements":[{"text":"Email"},{"text":"Password"}],"has_captcha":false,"has_error":false,"has_ban":false,"suggested_action":"fill_login_fields"}`,
		})
	}))
	defer server.Close()

	analyzer := NewCascadingScreenAnalyzer(config.Config{
		VLMBackends:    "ollama",
		OllamaURL:      server.URL,
		OllamaVLMModel: "qwen2.5vl:7b",
		VLMTimeoutSec:  2,
	}, http.DefaultClient, nil)

	got, err := analyzer.Analyze(context.Background(), "stub", "instagram", []byte("png"))
	if err != nil {
		t.Fatal(err)
	}
	if got.State != "login_screen" || got.BackendUsed != "ollama" || got.Confidence != 0.82 {
		t.Fatalf("unexpected analysis: %+v", got)
	}
	if len(got.Elements) != 2 || got.Elements[0] != "Email" {
		t.Fatalf("unexpected elements: %+v", got.Elements)
	}
}

func TestCascadingScreenAnalyzerFallsBackFromVisionServerToOllama(t *testing.T) {
	visionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	}))
	defer visionServer.Close()
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, map[string]any{
			"response": `{"screen_type":"feed","confidence":0.75,"description":"Feed is visible","elements":[{"text":"Home"}]}`,
		})
	}))
	defer ollamaServer.Close()

	analyzer := NewCascadingScreenAnalyzer(config.Config{
		VLMBackends:     "vision_server,ollama",
		VisionServerURL: visionServer.URL,
		OllamaURL:       ollamaServer.URL,
		OllamaVLMModel:  "qwen2.5vl:7b",
		VLMTimeoutSec:   2,
	}, http.DefaultClient, nil)

	got, err := analyzer.Analyze(context.Background(), "stub", "tiktok", []byte("png"))
	if err != nil {
		t.Fatal(err)
	}
	if got.State != "main_feed" || got.BackendUsed != "ollama" {
		t.Fatalf("unexpected fallback analysis: %+v", got)
	}
}

func TestCascadingScreenAnalyzerReturnsUnavailableWhenNoBackendsConfigured(t *testing.T) {
	analyzer := NewCascadingScreenAnalyzer(config.Config{VLMTimeoutSec: 1}, http.DefaultClient, nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := analyzer.Analyze(ctx, "stub", "android", []byte("png"))
	if err == nil {
		t.Fatal("expected error for missing VLM backends")
	}
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}
