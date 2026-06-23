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

func TestHTTPHandlerPostWaitForElementSuccess(t *testing.T) {
	dispatcher := &fakeObservationDispatcher{
		waitResult: domain.WaitForElementResult{
			Found:      true,
			FoundBy:    "text",
			WaitTimeMS: 3200,
			CheckCount: 7,
			Element: domain.UIElement{
				Type:        "Button",
				Text:        "OK",
				ResourceID:  "stub:id/ok",
				ContentDesc: "Create",
				Bounds:      domain.Bounds{X1: 0, Y1: 0, X2: 2, Y2: 2},
				Center:      domain.Point{X: 1, Y: 1},
			},
		},
	}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performWaitForElementRequest(h.Routes(), http.MethodPost, `{"serial":"stub","element":{"text":"OK","content_desc":"Create","match":"contains"},"priority":"high","timeout_sec":30,"check_interval_ms":500}`)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var got domain.WaitForElementResult
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !got.Found || got.FoundBy != "text" || got.WaitTimeMS != 3200 || got.CheckCount != 7 {
		t.Fatalf("unexpected response: %+v", got)
	}
	if dispatcher.waitSerial != "stub" || dispatcher.waitPriority != service.PriorityHigh {
		t.Fatalf("unexpected dispatcher call: serial=%q priority=%q", dispatcher.waitSerial, dispatcher.waitPriority)
	}
	if dispatcher.waitTimeout != 30*time.Second || dispatcher.waitCheckInterval != 500*time.Millisecond {
		t.Fatalf("unexpected durations: timeout=%s interval=%s", dispatcher.waitTimeout, dispatcher.waitCheckInterval)
	}
}

func TestHTTPHandlerPostWaitForElementTimeoutReturns408(t *testing.T) {
	dispatcher := &fakeObservationDispatcher{
		waitErr: domain.ErrElementWaitTimeout,
		waitResult: domain.WaitForElementResult{
			Found:      false,
			TimeoutSec: 30,
			WaitTimeMS: 30000,
			CheckCount: 60,
		},
	}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performWaitForElementRequest(h.Routes(), http.MethodPost, `{"serial":"stub","element":{"text":"Missing"},"timeout_sec":30,"check_interval_ms":500}`)
	if res.Code != http.StatusRequestTimeout {
		t.Fatalf("expected 408, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"error":"element_wait_timeout"`) {
		t.Fatalf("unexpected body: %s", res.Body.String())
	}
}

func TestHTTPHandlerPostWaitForElementValidationAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		err  error
		code int
	}{
		{name: "invalid json", body: `{`, code: http.StatusBadRequest},
		{name: "empty serial", body: `{"serial":"","element":{"text":"OK"}}`, code: http.StatusBadRequest},
		{name: "invalid serial", body: `{"serial":"bad serial","element":{"text":"OK"}}`, code: http.StatusBadRequest},
		{name: "empty element", body: `{"serial":"stub","element":{}}`, code: http.StatusBadRequest},
		{name: "invalid match", body: `{"serial":"stub","element":{"text":"OK","match":"fuzzy"}}`, code: http.StatusBadRequest},
		{name: "invalid priority", body: `{"serial":"stub","element":{"text":"OK"},"priority":"urgent"}`, code: http.StatusBadRequest},
		{name: "invalid timeout", body: `{"serial":"stub","element":{"text":"OK"},"timeout_sec":0}`, code: http.StatusBadRequest},
		{name: "invalid interval", body: `{"serial":"stub","element":{"text":"OK"},"check_interval_ms":99}`, code: http.StatusBadRequest},
		{name: "queue full", body: `{"serial":"stub","element":{"text":"OK"}}`, err: service.ErrQueueFull, code: http.StatusTooManyRequests},
		{name: "context timeout", body: `{"serial":"stub","element":{"text":"OK"}}`, err: context.DeadlineExceeded, code: http.StatusGatewayTimeout},
		{name: "dump failed", body: `{"serial":"stub","element":{"text":"OK"}}`, err: domain.ErrUIDumpFailed, code: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &fakeObservationDispatcher{waitErr: tc.err}
			h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

			res := performWaitForElementRequest(h.Routes(), http.MethodPost, tc.body)
			if res.Code != tc.code {
				t.Fatalf("expected %d, got %d: %s", tc.code, res.Code, res.Body.String())
			}
			if tc.err == nil && dispatcher.waitCalls != 0 {
				t.Fatalf("dispatcher called %d times", dispatcher.waitCalls)
			}
		})
	}
}

func TestHTTPHandlerPostWaitForElementMethodNotAllowed(t *testing.T) {
	dispatcher := &fakeObservationDispatcher{}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performWaitForElementRequest(h.Routes(), http.MethodGet, "")
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.Code)
	}
}

func performWaitForElementRequest(h http.Handler, method, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/wait-for-element", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}
