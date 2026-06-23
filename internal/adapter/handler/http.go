package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
	"github.com/mobilefarm/af/phone-observer/internal/port"
	"github.com/mobilefarm/af/phone-observer/internal/service"
)

var serialPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)

type observationDispatcher interface {
	Capture(ctx context.Context, serial string, priority service.ScreenshotPriority) (domain.Screenshot, error)
	DumpUI(ctx context.Context, serial string, priority service.ScreenshotPriority) (domain.UIDump, error)
	FindElement(ctx context.Context, serial string, query domain.FindElementQuery, priority service.ScreenshotPriority) (domain.FindElementResult, error)
	WaitForElement(ctx context.Context, serial string, query domain.FindElementQuery, priority service.ScreenshotPriority, waitTimeout, checkInterval time.Duration) (domain.WaitForElementResult, error)
	DetectState(ctx context.Context, serial string, options domain.DetectStateOptions, priority service.ScreenshotPriority) (domain.ScreenDetection, error)
	CurrentScreen(ctx context.Context, serial string, priority service.ScreenshotPriority) (domain.Screenshot, error)
	CurrentUI(ctx context.Context, serial string, priority service.ScreenshotPriority) (domain.UIDump, error)
	ClearCache(ctx context.Context, serial string, priority service.ScreenshotPriority) (domain.CacheClearResult, error)
}

type HTTPHandler struct {
	storage                  port.ObjectStorage
	observations             observationDispatcher
	defaultScreenshotTimeout time.Duration
	defaultDumpUITimeout     time.Duration
}

func NewHTTPHandler(storage port.ObjectStorage, observations observationDispatcher, defaultScreenshotTimeout, defaultDumpUITimeout time.Duration, _ port.Logger) *HTTPHandler {
	if defaultScreenshotTimeout <= 0 {
		defaultScreenshotTimeout = 10 * time.Second
	}
	if defaultDumpUITimeout <= 0 {
		defaultDumpUITimeout = 30 * time.Second
	}
	return &HTTPHandler{
		storage:                  storage,
		observations:             observations,
		defaultScreenshotTimeout: defaultScreenshotTimeout,
		defaultDumpUITimeout:     defaultDumpUITimeout,
	}
}

func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/ready", h.ready)
	mux.HandleFunc("/screenshot", h.screenshot)
	mux.HandleFunc("/dump-ui", h.dumpUI)
	mux.HandleFunc("/find-element", h.findElement)
	mux.HandleFunc("/wait-for-element", h.waitForElement)
	mux.HandleFunc("/detect-state", h.detectState)
	mux.HandleFunc("/screen/", h.currentScreen)
	mux.HandleFunc("/ui/", h.currentUI)
	mux.HandleFunc("/cache/", h.clearCache)
	return mux
}

func (h *HTTPHandler) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (h *HTTPHandler) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.storage.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready", "minio": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *HTTPHandler) screenshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}

	var req screenshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON"})
		return
	}
	if req.Serial == "" || !serialPattern.MatchString(req.Serial) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный serial"})
		return
	}
	if req.StoreInMinio != nil && !*req.StoreInMinio {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "store_in_minio=false не поддерживается"})
		return
	}

	priority, err := service.ParseScreenshotPriority(req.Priority)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный priority"})
		return
	}

	timeout := h.defaultScreenshotTimeout
	if req.TimeoutSec != nil {
		if *req.TimeoutSec <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "timeout_sec должен быть больше 0"})
			return
		}
		timeout = time.Duration(*req.TimeoutSec) * time.Second
	}
	if h.observations == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "screenshot dispatcher недоступен"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	shot, err := h.observations.Capture(ctx, req.Serial, priority)
	if err != nil {
		h.writeScreenshotError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, screenshotResponse{
		ScreenshotURL: shot.StorageURL,
		MinioKey:      shot.ObjectKey,
		SizeBytes:     shot.SizeBytes,
		Resolution: screenshotResolution{
			Width:  shot.Width,
			Height: shot.Height,
		},
		TakenAt: shot.TakenAt.Format(time.RFC3339),
	})
}

func (h *HTTPHandler) dumpUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}

	var req dumpUIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON"})
		return
	}
	if req.Serial == "" || !serialPattern.MatchString(req.Serial) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный serial"})
		return
	}

	format, err := parseDumpUIFormat(req.Format)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный format"})
		return
	}
	priority, err := service.ParseScreenshotPriority(req.Priority)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный priority"})
		return
	}

	timeout := h.defaultDumpUITimeout
	if req.TimeoutSec != nil {
		if *req.TimeoutSec <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "timeout_sec должен быть больше 0"})
			return
		}
		timeout = time.Duration(*req.TimeoutSec) * time.Second
	}
	if h.observations == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "observation dispatcher недоступен"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	dump, err := h.observations.DumpUI(ctx, req.Serial, priority)
	if err != nil {
		h.writeDumpUIError(w, err)
		return
	}

	resp := dumpUIResponse{
		ElementCount: dump.ElementCount,
		TakenAt:      dump.TakenAt.Format(time.RFC3339),
	}
	if format == dumpUIFormatXML {
		resp.XMLDump = dump.XMLDump
	} else {
		resp.Elements = dump.Elements
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) findElement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}

	var req findElementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON"})
		return
	}
	if req.Serial == "" || !serialPattern.MatchString(req.Serial) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный serial"})
		return
	}
	if err := domain.ValidateFindElementQuery(req.Element); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	priority, err := service.ParseScreenshotPriority(req.Priority)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный priority"})
		return
	}

	timeout := h.defaultDumpUITimeout
	if req.TimeoutSec != nil {
		if *req.TimeoutSec <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "timeout_sec должен быть больше 0"})
			return
		}
		timeout = time.Duration(*req.TimeoutSec) * time.Second
	}
	if h.observations == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "observation dispatcher недоступен"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	result, err := h.observations.FindElement(ctx, req.Serial, req.Element, priority)
	if err != nil {
		h.writeFindElementError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) waitForElement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}

	var req waitForElementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON"})
		return
	}
	if req.Serial == "" || !serialPattern.MatchString(req.Serial) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный serial"})
		return
	}
	if err := domain.ValidateFindElementQuery(req.Element); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	priority, err := service.ParseScreenshotPriority(req.Priority)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный priority"})
		return
	}

	waitTimeout := h.defaultDumpUITimeout
	if req.TimeoutSec != nil {
		if *req.TimeoutSec <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "timeout_sec должен быть больше 0"})
			return
		}
		waitTimeout = time.Duration(*req.TimeoutSec) * time.Second
	}
	checkInterval := 500 * time.Millisecond
	if req.CheckIntervalMS != nil {
		if *req.CheckIntervalMS < 100 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "check_interval_ms должен быть не меньше 100"})
			return
		}
		checkInterval = time.Duration(*req.CheckIntervalMS) * time.Millisecond
	}
	if h.observations == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "observation dispatcher недоступен"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), waitTimeout+h.defaultDumpUITimeout)
	defer cancel()

	result, err := h.observations.WaitForElement(ctx, req.Serial, req.Element, priority, waitTimeout, checkInterval)
	if err != nil {
		h.writeWaitForElementError(w, result, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) detectState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}

	var req detectStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON"})
		return
	}
	if req.Serial == "" || !serialPattern.MatchString(req.Serial) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный serial"})
		return
	}
	mode, err := domain.NormalizeDetectionMode(req.Mode)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	priority, err := service.ParseScreenshotPriority(req.Priority)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный priority"})
		return
	}

	useScreenshot := true
	if req.UseScreenshot != nil {
		useScreenshot = *req.UseScreenshot
	}
	storeScreenshot := false
	if req.StoreScreenshot != nil {
		storeScreenshot = *req.StoreScreenshot
	}
	if mode == domain.DetectionModeVLM && !useScreenshot {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mode=vlm требует use_screenshot=true"})
		return
	}

	timeout := h.defaultDumpUITimeout
	if req.TimeoutSec != nil {
		if *req.TimeoutSec <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "timeout_sec должен быть больше 0"})
			return
		}
		timeout = time.Duration(*req.TimeoutSec) * time.Second
	}
	if h.observations == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "observation dispatcher недоступен"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	result, err := h.observations.DetectState(ctx, req.Serial, domain.DetectStateOptions{
		Mode:            mode,
		Platform:        domain.NormalizeDetectionPlatform(req.Platform),
		UseScreenshot:   useScreenshot,
		StoreScreenshot: storeScreenshot,
	}, priority)
	if err != nil {
		h.writeDetectStateError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) currentScreen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}
	serial, ok := serialFromPath(r.URL.Path, "/screen/")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный serial"})
		return
	}
	priority, err := service.ParseScreenshotPriority(r.URL.Query().Get("priority"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный priority"})
		return
	}
	timeout, err := timeoutFromQuery(r, h.defaultScreenshotTimeout)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "timeout_sec должен быть больше 0"})
		return
	}
	if h.observations == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "observation dispatcher недоступен"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	shot, err := h.observations.CurrentScreen(ctx, serial, priority)
	if err != nil {
		h.writeScreenshotError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, currentScreenResponse{
		Serial:        shot.Serial,
		ScreenshotURL: shot.StorageURL,
		MinioKey:      shot.ObjectKey,
		SizeBytes:     shot.SizeBytes,
		Resolution: screenshotResolution{
			Width:  shot.Width,
			Height: shot.Height,
		},
		TakenAt: shot.TakenAt.Format(time.RFC3339),
	})
}

func (h *HTTPHandler) currentUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}
	serial, ok := serialFromPath(r.URL.Path, "/ui/")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный serial"})
		return
	}
	format, err := parseDumpUIFormat(r.URL.Query().Get("format"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный format"})
		return
	}
	priority, err := service.ParseScreenshotPriority(r.URL.Query().Get("priority"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный priority"})
		return
	}
	timeout, err := timeoutFromQuery(r, h.defaultDumpUITimeout)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "timeout_sec должен быть больше 0"})
		return
	}
	if h.observations == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "observation dispatcher недоступен"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	dump, err := h.observations.CurrentUI(ctx, serial, priority)
	if err != nil {
		h.writeDumpUIError(w, err)
		return
	}
	resp := currentUIResponse{
		Serial:       dump.Serial,
		ElementCount: dump.ElementCount,
		PackageName:  dump.PackageName,
		TakenAt:      dump.TakenAt.Format(time.RFC3339),
	}
	if format == dumpUIFormatXML {
		resp.XMLDump = dump.XMLDump
	} else {
		resp.Elements = dump.Elements
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) clearCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
		return
	}
	serial, ok := serialFromPath(r.URL.Path, "/cache/")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный serial"})
		return
	}
	priority, err := service.ParseScreenshotPriority(r.URL.Query().Get("priority"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неизвестный priority"})
		return
	}
	timeout, err := timeoutFromQuery(r, 5*time.Second)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "timeout_sec должен быть больше 0"})
		return
	}
	if h.observations == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "observation dispatcher недоступен"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	result, err := h.observations.ClearCache(ctx, serial, priority)
	if err != nil {
		h.writeClearCacheError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) writeScreenshotError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrQueueFull):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "очередь для телефона заполнена"})
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "таймаут выполнения screenshot"})
	case errors.Is(err, service.ErrInvalidPriority), errors.Is(err, domain.ErrInvalidSerial):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrStorageUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrScreenshotFailed):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func (h *HTTPHandler) writeClearCacheError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrQueueFull):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "очередь для телефона заполнена"})
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "таймаут выполнения clear-cache"})
	case errors.Is(err, service.ErrInvalidPriority), errors.Is(err, domain.ErrInvalidSerial):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func (h *HTTPHandler) writeFindElementError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrQueueFull):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "очередь для телефона заполнена"})
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "таймаут выполнения find-element"})
	case errors.Is(err, domain.ErrElementNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "element_not_found", "found": false, "found_by": ""})
	case errors.Is(err, service.ErrInvalidPriority), errors.Is(err, domain.ErrInvalidSerial), errors.Is(err, domain.ErrInvalidElementQuery):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrUIDumpFailed):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func (h *HTTPHandler) writeWaitForElementError(w http.ResponseWriter, result domain.WaitForElementResult, err error) {
	switch {
	case errors.Is(err, service.ErrQueueFull):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "очередь для телефона заполнена"})
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "таймаут выполнения wait-for-element"})
	case errors.Is(err, domain.ErrElementWaitTimeout):
		result.Error = "element_wait_timeout"
		result.Found = false
		result.FoundBy = ""
		writeJSON(w, http.StatusRequestTimeout, result)
	case errors.Is(err, service.ErrInvalidPriority), errors.Is(err, domain.ErrInvalidSerial), errors.Is(err, domain.ErrInvalidElementQuery):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrUIDumpFailed):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func (h *HTTPHandler) writeDetectStateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrQueueFull):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "очередь для телефона заполнена"})
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "таймаут выполнения detect-state"})
	case errors.Is(err, domain.ErrVLMUnavailable), errors.Is(err, domain.ErrStorageUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidPriority), errors.Is(err, domain.ErrInvalidSerial), errors.Is(err, domain.ErrInvalidDetectMode):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrUIDumpFailed), errors.Is(err, domain.ErrScreenshotFailed), errors.Is(err, domain.ErrVLMFailed):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func (h *HTTPHandler) writeDumpUIError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrQueueFull):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "очередь для телефона заполнена"})
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "таймаут выполнения dump-ui"})
	case errors.Is(err, service.ErrInvalidPriority), errors.Is(err, domain.ErrInvalidSerial):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrUIDumpFailed):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

type screenshotRequest struct {
	Serial       string `json:"serial"`
	StoreInMinio *bool  `json:"store_in_minio"`
	TimeoutSec   *int   `json:"timeout_sec"`
	Priority     string `json:"priority"`
}

type screenshotResponse struct {
	ScreenshotURL string               `json:"screenshot_url"`
	MinioKey      string               `json:"minio_key"`
	SizeBytes     int64                `json:"size_bytes"`
	Resolution    screenshotResolution `json:"resolution"`
	TakenAt       string               `json:"taken_at"`
}

type currentScreenResponse struct {
	Serial        string               `json:"serial"`
	ScreenshotURL string               `json:"screenshot_url"`
	MinioKey      string               `json:"minio_key"`
	SizeBytes     int64                `json:"size_bytes"`
	Resolution    screenshotResolution `json:"resolution"`
	TakenAt       string               `json:"taken_at"`
}

type screenshotResolution struct {
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
}

const (
	dumpUIFormatJSON = "json"
	dumpUIFormatXML  = "xml"
)

type dumpUIRequest struct {
	Serial     string `json:"serial"`
	Format     string `json:"format"`
	TimeoutSec *int   `json:"timeout_sec"`
	Priority   string `json:"priority"`
}

type dumpUIResponse struct {
	Elements     []domain.UIElement `json:"elements,omitempty"`
	XMLDump      string             `json:"xml_dump,omitempty"`
	ElementCount int                `json:"element_count"`
	TakenAt      string             `json:"taken_at"`
}

type currentUIResponse struct {
	Serial       string             `json:"serial"`
	Elements     []domain.UIElement `json:"elements,omitempty"`
	XMLDump      string             `json:"xml_dump,omitempty"`
	ElementCount int                `json:"element_count"`
	PackageName  string             `json:"package_name"`
	TakenAt      string             `json:"taken_at"`
}

type findElementRequest struct {
	Serial     string                  `json:"serial"`
	Element    domain.FindElementQuery `json:"element"`
	TimeoutSec *int                    `json:"timeout_sec"`
	Priority   string                  `json:"priority"`
}

type waitForElementRequest struct {
	Serial          string                  `json:"serial"`
	Element         domain.FindElementQuery `json:"element"`
	TimeoutSec      *int                    `json:"timeout_sec"`
	CheckIntervalMS *int                    `json:"check_interval_ms"`
	Priority        string                  `json:"priority"`
}

type detectStateRequest struct {
	Serial          string `json:"serial"`
	Mode            string `json:"mode"`
	Platform        string `json:"platform"`
	UseScreenshot   *bool  `json:"use_screenshot"`
	StoreScreenshot *bool  `json:"store_screenshot"`
	TimeoutSec      *int   `json:"timeout_sec"`
	Priority        string `json:"priority"`
}

func parseDumpUIFormat(raw string) (string, error) {
	switch raw {
	case "", dumpUIFormatJSON:
		return dumpUIFormatJSON, nil
	case dumpUIFormatXML:
		return dumpUIFormatXML, nil
	default:
		return "", errors.New("неизвестный format")
	}
}

func serialFromPath(path, prefix string) (string, bool) {
	serial := strings.TrimPrefix(path, prefix)
	if serial == path || serial == "" || strings.Contains(serial, "/") || !serialPattern.MatchString(serial) {
		return "", false
	}
	return serial, true
}

func timeoutFromQuery(r *http.Request, fallback time.Duration) (time.Duration, error) {
	raw := r.URL.Query().Get("timeout_sec")
	if raw == "" {
		return fallback, nil
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 0, errors.New("timeout_sec должен быть больше 0")
	}
	return time.Duration(seconds) * time.Second, nil
}
