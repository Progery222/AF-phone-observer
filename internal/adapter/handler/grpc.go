package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/port"
	"github.com/mobilefarm/af/phone-observer/internal/service"
	"google.golang.org/grpc"
)

type ObserverHandler struct {
	observe    *service.ObserveService
	screenshot *service.ScreenshotService
	log        port.Logger
}

func NewObserverHandler(observe *service.ObserveService, screenshot *service.ScreenshotService, log port.Logger) *ObserverHandler {
	return &ObserverHandler{observe: observe, screenshot: screenshot, log: log}
}

func (h *ObserverHandler) Register(s *grpc.Server) {
	_ = s // observerpb.RegisterObserverServiceServer(s, h)
}

type HealthHandler struct {
	storage port.ObjectStorage
}

func NewHealthHandler(storage port.ObjectStorage) *HealthHandler {
	return &HealthHandler{storage: storage}
}

func (h *HealthHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ready", h.ready)
	return mux
}

func (h *HealthHandler) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.storage.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready", "minio": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(v)
}
