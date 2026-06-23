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

func TestHTTPHandlerGetScreenSuccess(t *testing.T) {
	takenAt := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	dispatcher := &fakeObservationDispatcher{
		currentScreenShot: domain.Screenshot{
			Serial:     "R5GL2218DMR",
			ObjectKey:  "R5GL2218DMR/20260623-090000.png",
			StorageURL: "af-screenshots/R5GL2218DMR/20260623-090000.png",
			SizeBytes:  245760,
			Width:      1080,
			Height:     1920,
			TakenAt:    takenAt,
		},
	}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performCacheRequest(h.Routes(), http.MethodGet, "/screen/R5GL2218DMR?priority=high&timeout_sec=12")
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var got struct {
		Serial        string `json:"serial"`
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
	if got.Serial != "R5GL2218DMR" || got.ScreenshotURL == "" || got.MinioKey == "" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if got.Resolution.Width != 1080 || got.Resolution.Height != 1920 || got.TakenAt != "2026-06-23T09:00:00Z" {
		t.Fatalf("unexpected metadata: %+v", got)
	}
	if dispatcher.currentScreenSerial != "R5GL2218DMR" || dispatcher.currentScreenPriority != service.PriorityHigh {
		t.Fatalf("unexpected dispatcher call: serial=%q priority=%q", dispatcher.currentScreenSerial, dispatcher.currentScreenPriority)
	}
}

func TestHTTPHandlerGetUISuccessJSONAndXML(t *testing.T) {
	for _, tc := range []struct {
		name      string
		path      string
		wantXML   bool
		wantPrio  service.ScreenshotPriority
		wantCount int
	}{
		{name: "json default", path: "/ui/stub", wantPrio: service.PriorityNormal, wantCount: 1},
		{name: "xml high", path: "/ui/stub?format=xml&priority=high&timeout_sec=30", wantXML: true, wantPrio: service.PriorityHigh, wantCount: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &fakeObservationDispatcher{
				currentUIDump: domain.UIDump{
					Serial:       "stub",
					XMLDump:      `<hierarchy><node class="android.widget.Button" text="OK" bounds="[0,0][2,2]" /></hierarchy>`,
					ElementCount: 1,
					PackageName:  "com.example",
					TakenAt:      time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC),
					Elements: []domain.UIElement{
						{Type: "Button", Text: "OK", Bounds: domain.Bounds{X1: 0, Y1: 0, X2: 2, Y2: 2}, Center: domain.Point{X: 1, Y: 1}},
					},
				},
			}
			h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

			res := performCacheRequest(h.Routes(), http.MethodGet, tc.path)
			if res.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
			}
			var got struct {
				Serial       string             `json:"serial"`
				Elements     []domain.UIElement `json:"elements"`
				XMLDump      string             `json:"xml_dump"`
				ElementCount int                `json:"element_count"`
				PackageName  string             `json:"package_name"`
				TakenAt      string             `json:"taken_at"`
			}
			if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got.Serial != "stub" || got.ElementCount != tc.wantCount || got.PackageName != "com.example" {
				t.Fatalf("unexpected response: %+v", got)
			}
			if tc.wantXML && got.XMLDump == "" {
				t.Fatalf("expected xml dump: %+v", got)
			}
			if !tc.wantXML && len(got.Elements) != 1 {
				t.Fatalf("expected json elements: %+v", got)
			}
			if dispatcher.currentUISerial != "stub" || dispatcher.currentUIPriority != tc.wantPrio {
				t.Fatalf("unexpected dispatcher call: serial=%q priority=%q", dispatcher.currentUISerial, dispatcher.currentUIPriority)
			}
		})
	}
}

func TestHTTPHandlerDeleteCacheSuccess(t *testing.T) {
	takenAt := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	dispatcher := &fakeObservationDispatcher{
		clearResult: domain.CacheClearResult{
			Serial:        "stub",
			Cleared:       true,
			ScreenCleared: true,
			UICleared:     true,
			ClearedAt:     takenAt,
		},
	}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performCacheRequest(h.Routes(), http.MethodDelete, "/cache/stub?priority=high&timeout_sec=5")
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var got domain.CacheClearResult
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !got.Cleared || !got.ScreenCleared || !got.UICleared || got.Serial != "stub" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if dispatcher.clearSerial != "stub" || dispatcher.clearPriority != service.PriorityHigh {
		t.Fatalf("unexpected dispatcher call: serial=%q priority=%q", dispatcher.clearSerial, dispatcher.clearPriority)
	}
}

func TestHTTPHandlerScreenCacheValidationAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name       string
		method     string
		path       string
		screenErr  error
		uiErr      error
		clearErr   error
		wantStatus int
	}{
		{name: "screen invalid serial", method: http.MethodGet, path: "/screen/bad/serial", wantStatus: http.StatusBadRequest},
		{name: "screen invalid priority", method: http.MethodGet, path: "/screen/stub?priority=urgent", wantStatus: http.StatusBadRequest},
		{name: "screen invalid timeout", method: http.MethodGet, path: "/screen/stub?timeout_sec=0", wantStatus: http.StatusBadRequest},
		{name: "screen queue full", method: http.MethodGet, path: "/screen/stub", screenErr: service.ErrQueueFull, wantStatus: http.StatusTooManyRequests},
		{name: "screen storage", method: http.MethodGet, path: "/screen/stub", screenErr: domain.ErrStorageUnavailable, wantStatus: http.StatusServiceUnavailable},
		{name: "screen timeout", method: http.MethodGet, path: "/screen/stub", screenErr: context.DeadlineExceeded, wantStatus: http.StatusGatewayTimeout},
		{name: "ui invalid format", method: http.MethodGet, path: "/ui/stub?format=yaml", wantStatus: http.StatusBadRequest},
		{name: "ui dump failed", method: http.MethodGet, path: "/ui/stub", uiErr: domain.ErrUIDumpFailed, wantStatus: http.StatusInternalServerError},
		{name: "cache queue full", method: http.MethodDelete, path: "/cache/stub", clearErr: service.ErrQueueFull, wantStatus: http.StatusTooManyRequests},
		{name: "cache invalid method", method: http.MethodGet, path: "/cache/stub", wantStatus: http.StatusMethodNotAllowed},
		{name: "screen invalid method", method: http.MethodPost, path: "/screen/stub", wantStatus: http.StatusMethodNotAllowed},
		{name: "ui invalid method", method: http.MethodPost, path: "/ui/stub", wantStatus: http.StatusMethodNotAllowed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &fakeObservationDispatcher{
				currentScreenErr: tc.screenErr,
				currentUIErr:     tc.uiErr,
				clearErr:         tc.clearErr,
			}
			h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

			res := performCacheRequest(h.Routes(), tc.method, tc.path)
			if res.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tc.wantStatus, res.Code, res.Body.String())
			}
		})
	}
}

func performCacheRequest(h http.Handler, method, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(""))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}
