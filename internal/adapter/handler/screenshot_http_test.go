package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
	"github.com/mobilefarm/af/phone-observer/internal/service"
)

type fakeScreenshotDispatcher struct {
	mu       sync.Mutex
	serial   string
	priority service.ScreenshotPriority
	calls    int
	shot     domain.Screenshot
	err      error
}

func (f *fakeScreenshotDispatcher) Capture(_ context.Context, serial string, priority service.ScreenshotPriority) (domain.Screenshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.serial = serial
	f.priority = priority
	if f.err != nil {
		return domain.Screenshot{}, f.err
	}
	return f.shot, nil
}

func (f *fakeScreenshotDispatcher) DumpUI(context.Context, string, service.ScreenshotPriority) (domain.UIDump, error) {
	return domain.UIDump{}, nil
}

func (f *fakeScreenshotDispatcher) FindElement(context.Context, string, domain.FindElementQuery, service.ScreenshotPriority) (domain.FindElementResult, error) {
	return domain.FindElementResult{}, nil
}

func (f *fakeScreenshotDispatcher) WaitForElement(context.Context, string, domain.FindElementQuery, service.ScreenshotPriority, time.Duration, time.Duration) (domain.WaitForElementResult, error) {
	return domain.WaitForElementResult{}, nil
}

func (f *fakeScreenshotDispatcher) DetectState(context.Context, string, domain.DetectStateOptions, service.ScreenshotPriority) (domain.ScreenDetection, error) {
	return domain.ScreenDetection{}, nil
}

func (f *fakeScreenshotDispatcher) CurrentScreen(ctx context.Context, serial string, priority service.ScreenshotPriority) (domain.Screenshot, error) {
	return f.Capture(ctx, serial, priority)
}

func (f *fakeScreenshotDispatcher) CurrentUI(context.Context, string, service.ScreenshotPriority) (domain.UIDump, error) {
	return domain.UIDump{}, nil
}

func (f *fakeScreenshotDispatcher) ClearCache(context.Context, string, service.ScreenshotPriority) (domain.CacheClearResult, error) {
	return domain.CacheClearResult{}, nil
}

type fakeReadyStorage struct{}

func (fakeReadyStorage) Upload(context.Context, string, []byte) (string, error) { return "", nil }
func (fakeReadyStorage) Ping(context.Context) error                             { return nil }

func TestHTTPHandlerPostScreenshotSuccess(t *testing.T) {
	for _, tc := range []struct {
		name     string
		body     string
		priority service.ScreenshotPriority
	}{
		{
			name:     "normal",
			body:     `{"serial":"R83YA05Y51P","store_in_minio":true,"timeout_sec":10,"priority":"normal"}`,
			priority: service.PriorityNormal,
		},
		{
			name:     "high",
			body:     `{"serial":"R83YA05Y51P","store_in_minio":true,"timeout_sec":10,"priority":"high"}`,
			priority: service.PriorityHigh,
		},
		{
			name:     "default priority",
			body:     `{"serial":"R83YA05Y51P","store_in_minio":true,"timeout_sec":10}`,
			priority: service.PriorityNormal,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			takenAt := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
			dispatcher := &fakeScreenshotDispatcher{
				shot: domain.Screenshot{
					Serial:     "R83YA05Y51P",
					ObjectKey:  "R83YA05Y51P/20260623-120000.png",
					StorageURL: "af-screenshots/R83YA05Y51P/20260623-120000.png",
					SizeBytes:  245760,
					Width:      1080,
					Height:     1920,
					TakenAt:    takenAt,
				},
			}
			h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

			res := performScreenshotRequest(h.Routes(), http.MethodPost, tc.body)
			if res.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
			}

			var got struct {
				ScreenshotURL string `json:"screenshot_url"`
				MinioKey      string `json:"minio_key"`
				SizeBytes     int64  `json:"size_bytes"`
				Resolution    struct {
					Width  int32 `json:"width"`
					Height int32 `json:"height"`
				} `json:"resolution"`
				TakenAt string `json:"taken_at"`
			}
			if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got.ScreenshotURL != "af-screenshots/R83YA05Y51P/20260623-120000.png" {
				t.Fatalf("unexpected screenshot_url: %q", got.ScreenshotURL)
			}
			if got.MinioKey != "R83YA05Y51P/20260623-120000.png" {
				t.Fatalf("unexpected minio_key: %q", got.MinioKey)
			}
			if got.SizeBytes != 245760 || got.Resolution.Width != 1080 || got.Resolution.Height != 1920 {
				t.Fatalf("unexpected metadata: %+v", got)
			}
			if got.TakenAt != "2026-06-23T09:00:00Z" {
				t.Fatalf("unexpected taken_at: %q", got.TakenAt)
			}
			if dispatcher.serial != "R83YA05Y51P" || dispatcher.priority != tc.priority {
				t.Fatalf("unexpected dispatcher call: serial=%q priority=%q", dispatcher.serial, dispatcher.priority)
			}
		})
	}
}

func TestHTTPHandlerPostScreenshotValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		code int
	}{
		{name: "invalid json", body: `{`, code: http.StatusBadRequest},
		{name: "empty serial", body: `{"serial":"","store_in_minio":true}`, code: http.StatusBadRequest},
		{name: "invalid serial", body: `{"serial":"bad serial","store_in_minio":true}`, code: http.StatusBadRequest},
		{name: "invalid priority", body: `{"serial":"R83YA05Y51P","store_in_minio":true,"priority":"urgent"}`, code: http.StatusBadRequest},
		{name: "store disabled", body: `{"serial":"R83YA05Y51P","store_in_minio":false}`, code: http.StatusBadRequest},
		{name: "invalid timeout", body: `{"serial":"R83YA05Y51P","store_in_minio":true,"timeout_sec":0}`, code: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &fakeScreenshotDispatcher{}
			h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

			res := performScreenshotRequest(h.Routes(), http.MethodPost, tc.body)
			if res.Code != tc.code {
				t.Fatalf("expected %d, got %d: %s", tc.code, res.Code, res.Body.String())
			}
			if dispatcher.calls != 0 {
				t.Fatalf("dispatcher called %d times", dispatcher.calls)
			}
		})
	}
}

func TestHTTPHandlerPostScreenshotErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		code int
	}{
		{name: "queue full", err: service.ErrQueueFull, code: http.StatusTooManyRequests},
		{name: "timeout", err: context.DeadlineExceeded, code: http.StatusGatewayTimeout},
		{name: "storage", err: domain.ErrStorageUnavailable, code: http.StatusServiceUnavailable},
		{name: "capture", err: domain.ErrScreenshotFailed, code: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &fakeScreenshotDispatcher{err: tc.err}
			h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

			res := performScreenshotRequest(h.Routes(), http.MethodPost, `{"serial":"R83YA05Y51P","store_in_minio":true}`)
			if res.Code != tc.code {
				t.Fatalf("expected %d, got %d: %s", tc.code, res.Code, res.Body.String())
			}
		})
	}
}

func TestHTTPHandlerPostScreenshotMethodNotAllowed(t *testing.T) {
	dispatcher := &fakeScreenshotDispatcher{}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performScreenshotRequest(h.Routes(), http.MethodGet, "")
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.Code)
	}
	if dispatcher.calls != 0 {
		t.Fatalf("dispatcher called %d times", dispatcher.calls)
	}
}

func TestHTTPHandlerPostScreenshotQueueFullByWrappedError(t *testing.T) {
	dispatcher := &fakeScreenshotDispatcher{err: errors.Join(service.ErrQueueFull, context.Canceled)}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performScreenshotRequest(h.Routes(), http.MethodPost, `{"serial":"R83YA05Y51P","store_in_minio":true}`)
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d: %s", res.Code, res.Body.String())
	}
}

func performScreenshotRequest(h http.Handler, method, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/screenshot", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}
