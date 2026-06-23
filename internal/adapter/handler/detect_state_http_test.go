package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
	"github.com/mobilefarm/af/phone-observer/internal/service"
)

func TestHTTPHandlerPostDetectStateSuccess(t *testing.T) {
	takenAt := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	dispatcher := &fakeObservationDispatcher{
		detectResult: domain.ScreenDetection{
			State:           "login_screen",
			Confidence:      0.95,
			Source:          "hybrid",
			BackendUsed:     "ollama",
			Description:     "Экран входа",
			Elements:        []string{"Email", "Password", "Log in"},
			MatchedSignals:  []string{"ui:email_input", "vlm:login"},
			Flags:           domain.DetectionFlags{},
			SuggestedAction: "fill_login_fields",
			PackageName:     "com.instagram.android",
			ElementCount:    42,
			ScreenshotURL:   "af-screenshots/stub/20260623-090000.png",
			MinioKey:        "stub/20260623-090000.png",
			TakenAt:         takenAt,
		},
	}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performDetectStateRequest(h.Routes(), http.MethodPost, `{"serial":"stub","mode":"vlm","platform":"instagram","use_screenshot":true,"store_screenshot":true,"priority":"high","timeout_sec":30}`)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var got domain.ScreenDetection
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.State != "login_screen" || got.BackendUsed != "ollama" || got.ScreenshotURL == "" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if dispatcher.detectSerial != "stub" || dispatcher.detectPriority != service.PriorityHigh {
		t.Fatalf("unexpected dispatcher call: serial=%q priority=%q", dispatcher.detectSerial, dispatcher.detectPriority)
	}
	if dispatcher.detectOptions.Mode != domain.DetectionModeVLM || dispatcher.detectOptions.Platform != "instagram" || !dispatcher.detectOptions.UseScreenshot || !dispatcher.detectOptions.StoreScreenshot {
		t.Fatalf("unexpected options: %+v", dispatcher.detectOptions)
	}
}

func TestHTTPHandlerPostDetectStateUnknownIsOK(t *testing.T) {
	dispatcher := &fakeObservationDispatcher{
		detectResult: domain.ScreenDetection{
			State:       "unknown",
			Confidence:  0,
			Source:      "ui",
			Description: "Экран прочитан, но состояние не распознано",
			TakenAt:     time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC),
		},
	}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performDetectStateRequest(h.Routes(), http.MethodPost, `{"serial":"stub","mode":"ui"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"state":"unknown"`) {
		t.Fatalf("unexpected body: %s", res.Body.String())
	}
}

func TestHTTPHandlerPostDetectStateValidationAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		err  error
		code int
	}{
		{name: "invalid json", body: `{`, code: http.StatusBadRequest},
		{name: "empty serial", body: `{"serial":""}`, code: http.StatusBadRequest},
		{name: "invalid serial", body: `{"serial":"bad serial"}`, code: http.StatusBadRequest},
		{name: "invalid mode", body: `{"serial":"stub","mode":"magic"}`, code: http.StatusBadRequest},
		{name: "invalid priority", body: `{"serial":"stub","priority":"urgent"}`, code: http.StatusBadRequest},
		{name: "invalid timeout", body: `{"serial":"stub","timeout_sec":0}`, code: http.StatusBadRequest},
		{name: "vlm without screenshot", body: `{"serial":"stub","mode":"vlm","use_screenshot":false}`, code: http.StatusBadRequest},
		{name: "queue full", body: `{"serial":"stub","mode":"ui"}`, err: service.ErrQueueFull, code: http.StatusTooManyRequests},
		{name: "timeout", body: `{"serial":"stub","mode":"ui"}`, err: context.DeadlineExceeded, code: http.StatusGatewayTimeout},
		{name: "vlm unavailable", body: `{"serial":"stub","mode":"vlm"}`, err: domain.ErrVLMUnavailable, code: http.StatusServiceUnavailable},
		{name: "storage unavailable", body: `{"serial":"stub","store_screenshot":true}`, err: domain.ErrStorageUnavailable, code: http.StatusServiceUnavailable},
		{name: "dump failed", body: `{"serial":"stub","mode":"ui"}`, err: domain.ErrUIDumpFailed, code: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &fakeObservationDispatcher{detectErr: tc.err}
			h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

			res := performDetectStateRequest(h.Routes(), http.MethodPost, tc.body)
			if res.Code != tc.code {
				t.Fatalf("expected %d, got %d: %s", tc.code, res.Code, res.Body.String())
			}
			if tc.err == nil && dispatcher.detectCalls != 0 {
				t.Fatalf("dispatcher called %d times", dispatcher.detectCalls)
			}
		})
	}
}

func TestHTTPHandlerPostDetectStateMethodNotAllowed(t *testing.T) {
	dispatcher := &fakeObservationDispatcher{}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performDetectStateRequest(h.Routes(), http.MethodGet, "")
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.Code)
	}
}

func performDetectStateRequest(h http.Handler, method, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/detect-state", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}
