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

type fakeObservationDispatcher struct {
	captureSerial   string
	capturePriority service.ScreenshotPriority
	captureCalls    int
	captureShot     domain.Screenshot
	captureErr      error

	dumpSerial   string
	dumpPriority service.ScreenshotPriority
	dumpCalls    int
	dump         domain.UIDump
	dumpErr      error

	findSerial   string
	findPriority service.ScreenshotPriority
	findCalls    int
	findResult   domain.FindElementResult
	findErr      error

	waitSerial        string
	waitPriority      service.ScreenshotPriority
	waitTimeout       time.Duration
	waitCheckInterval time.Duration
	waitCalls         int
	waitResult        domain.WaitForElementResult
	waitErr           error

	detectSerial   string
	detectPriority service.ScreenshotPriority
	detectOptions  domain.DetectStateOptions
	detectCalls    int
	detectResult   domain.ScreenDetection
	detectErr      error

	currentScreenSerial   string
	currentScreenPriority service.ScreenshotPriority
	currentScreenCalls    int
	currentScreenShot     domain.Screenshot
	currentScreenErr      error

	currentUISerial   string
	currentUIPriority service.ScreenshotPriority
	currentUICalls    int
	currentUIDump     domain.UIDump
	currentUIErr      error

	clearSerial   string
	clearPriority service.ScreenshotPriority
	clearCalls    int
	clearResult   domain.CacheClearResult
	clearErr      error
}

func (f *fakeObservationDispatcher) Capture(_ context.Context, serial string, priority service.ScreenshotPriority) (domain.Screenshot, error) {
	f.captureCalls++
	f.captureSerial = serial
	f.capturePriority = priority
	if f.captureErr != nil {
		return domain.Screenshot{}, f.captureErr
	}
	return f.captureShot, nil
}

func (f *fakeObservationDispatcher) DumpUI(_ context.Context, serial string, priority service.ScreenshotPriority) (domain.UIDump, error) {
	f.dumpCalls++
	f.dumpSerial = serial
	f.dumpPriority = priority
	if f.dumpErr != nil {
		return domain.UIDump{}, f.dumpErr
	}
	return f.dump, nil
}

func (f *fakeObservationDispatcher) FindElement(_ context.Context, serial string, query domain.FindElementQuery, priority service.ScreenshotPriority) (domain.FindElementResult, error) {
	f.findCalls++
	f.findSerial = serial
	f.findPriority = priority
	if f.findErr != nil {
		return domain.FindElementResult{}, f.findErr
	}
	return f.findResult, nil
}

func (f *fakeObservationDispatcher) WaitForElement(_ context.Context, serial string, query domain.FindElementQuery, priority service.ScreenshotPriority, waitTimeout, checkInterval time.Duration) (domain.WaitForElementResult, error) {
	f.waitCalls++
	f.waitSerial = serial
	f.waitPriority = priority
	f.waitTimeout = waitTimeout
	f.waitCheckInterval = checkInterval
	if f.waitErr != nil {
		return domain.WaitForElementResult{}, f.waitErr
	}
	return f.waitResult, nil
}

func (f *fakeObservationDispatcher) DetectState(_ context.Context, serial string, options domain.DetectStateOptions, priority service.ScreenshotPriority) (domain.ScreenDetection, error) {
	f.detectCalls++
	f.detectSerial = serial
	f.detectOptions = options
	f.detectPriority = priority
	if f.detectErr != nil {
		return domain.ScreenDetection{}, f.detectErr
	}
	return f.detectResult, nil
}

func (f *fakeObservationDispatcher) CurrentScreen(_ context.Context, serial string, priority service.ScreenshotPriority) (domain.Screenshot, error) {
	f.currentScreenCalls++
	f.currentScreenSerial = serial
	f.currentScreenPriority = priority
	if f.currentScreenErr != nil {
		return domain.Screenshot{}, f.currentScreenErr
	}
	return f.currentScreenShot, nil
}

func (f *fakeObservationDispatcher) CurrentUI(_ context.Context, serial string, priority service.ScreenshotPriority) (domain.UIDump, error) {
	f.currentUICalls++
	f.currentUISerial = serial
	f.currentUIPriority = priority
	if f.currentUIErr != nil {
		return domain.UIDump{}, f.currentUIErr
	}
	return f.currentUIDump, nil
}

func (f *fakeObservationDispatcher) ClearCache(_ context.Context, serial string, priority service.ScreenshotPriority) (domain.CacheClearResult, error) {
	f.clearCalls++
	f.clearSerial = serial
	f.clearPriority = priority
	if f.clearErr != nil {
		return domain.CacheClearResult{}, f.clearErr
	}
	return f.clearResult, nil
}

func TestHTTPHandlerPostDumpUISuccessJSON(t *testing.T) {
	takenAt := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	dispatcher := &fakeObservationDispatcher{
		dump: domain.UIDump{
			Serial:       "R83YA05Y51P",
			XMLDump:      "<hierarchy/>",
			ElementCount: 1,
			TakenAt:      takenAt,
			Elements: []domain.UIElement{
				{
					Type:       "Button",
					Text:       "Войти",
					ResourceID: "com.app:id/login",
					Bounds:     domain.Bounds{X1: 200, Y1: 500, X2: 600, Y2: 580},
					Center:     domain.Point{X: 400, Y: 540},
				},
			},
		},
	}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performDumpUIRequest(h.Routes(), http.MethodPost, `{"serial":"R83YA05Y51P","format":"json","timeout_sec":30,"priority":"high"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var got struct {
		Elements []struct {
			Type       string        `json:"type"`
			Text       string        `json:"text"`
			ResourceID string        `json:"resource_id"`
			Hint       string        `json:"hint"`
			Bounds     domain.Bounds `json:"bounds"`
			Center     domain.Point  `json:"center"`
		} `json:"elements"`
		ElementCount int    `json:"element_count"`
		TakenAt      string `json:"taken_at"`
	}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ElementCount != 1 || len(got.Elements) != 1 {
		t.Fatalf("unexpected element count: %+v", got)
	}
	if got.Elements[0].Type != "Button" || got.Elements[0].Center.X != 400 {
		t.Fatalf("unexpected element: %+v", got.Elements[0])
	}
	if got.TakenAt != "2026-06-23T09:00:00Z" {
		t.Fatalf("unexpected taken_at: %q", got.TakenAt)
	}
	if dispatcher.dumpSerial != "R83YA05Y51P" || dispatcher.dumpPriority != service.PriorityHigh {
		t.Fatalf("unexpected dispatcher call: serial=%q priority=%q", dispatcher.dumpSerial, dispatcher.dumpPriority)
	}
}

func TestHTTPHandlerPostDumpUISuccessXMLAndDefaultFormat(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "xml", body: `{"serial":"stub","format":"xml"}`},
		{name: "default", body: `{"serial":"stub"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &fakeObservationDispatcher{
				dump: domain.UIDump{
					Serial:       "stub",
					XMLDump:      `<hierarchy><node class="android.widget.Button" bounds="[0,0][2,2]" /></hierarchy>`,
					ElementCount: 1,
					TakenAt:      time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC),
				},
			}
			h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

			res := performDumpUIRequest(h.Routes(), http.MethodPost, tc.body)
			if res.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
			}
			if dispatcher.dumpPriority != service.PriorityNormal {
				t.Fatalf("expected default normal priority, got %q", dispatcher.dumpPriority)
			}
		})
	}
}

func TestHTTPHandlerPostDumpUIValidationAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		err  error
		code int
	}{
		{name: "invalid json", body: `{`, code: http.StatusBadRequest},
		{name: "empty serial", body: `{"serial":""}`, code: http.StatusBadRequest},
		{name: "invalid serial", body: `{"serial":"bad serial"}`, code: http.StatusBadRequest},
		{name: "invalid format", body: `{"serial":"stub","format":"yaml"}`, code: http.StatusBadRequest},
		{name: "invalid priority", body: `{"serial":"stub","priority":"urgent"}`, code: http.StatusBadRequest},
		{name: "invalid timeout", body: `{"serial":"stub","timeout_sec":0}`, code: http.StatusBadRequest},
		{name: "queue full", body: `{"serial":"stub"}`, err: service.ErrQueueFull, code: http.StatusTooManyRequests},
		{name: "timeout", body: `{"serial":"stub"}`, err: context.DeadlineExceeded, code: http.StatusGatewayTimeout},
		{name: "dump failed", body: `{"serial":"stub"}`, err: domain.ErrUIDumpFailed, code: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &fakeObservationDispatcher{dumpErr: tc.err}
			h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

			res := performDumpUIRequest(h.Routes(), http.MethodPost, tc.body)
			if res.Code != tc.code {
				t.Fatalf("expected %d, got %d: %s", tc.code, res.Code, res.Body.String())
			}
			if tc.err == nil && dispatcher.dumpCalls != 0 {
				t.Fatalf("dispatcher called %d times", dispatcher.dumpCalls)
			}
		})
	}
}

func TestHTTPHandlerPostDumpUIMethodNotAllowed(t *testing.T) {
	dispatcher := &fakeObservationDispatcher{}
	h := NewHTTPHandler(fakeReadyStorage{}, dispatcher, 10*time.Second, 30*time.Second, slog.Default())

	res := performDumpUIRequest(h.Routes(), http.MethodGet, "")
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.Code)
	}
	if dispatcher.dumpCalls != 0 {
		t.Fatalf("dispatcher called %d times", dispatcher.dumpCalls)
	}
}

func performDumpUIRequest(h http.Handler, method, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/dump-ui", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}
