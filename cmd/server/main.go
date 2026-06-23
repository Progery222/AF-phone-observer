package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/adapter/driver"
	"github.com/mobilefarm/af/phone-observer/internal/adapter/handler"
	"github.com/mobilefarm/af/phone-observer/internal/adapter/repository"
	"github.com/mobilefarm/af/phone-observer/internal/config"
	"github.com/mobilefarm/af/phone-observer/internal/port"
	"github.com/mobilefarm/af/phone-observer/internal/service"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ui := driver.NewUIAutomatorDriver(logger)
	shot := driver.NewAdbScreenshotDriver(logger)

	var storage port.ObjectStorage
	storage, closeStorage, err := repository.NewMinIOStorage(cfg)
	if err != nil {
		logger.Warn("minio unavailable, storage disabled", "error", err)
		storage = repository.NewNoopStorage()
		closeStorage = func() {}
	}
	defer closeStorage()

	observeSvc := service.NewObserveService(ui, logger)
	screenshotSvc := service.NewScreenshotService(shot, storage, logger)
	observationDispatcher := service.NewObservationDispatcher(
		screenshotSvc,
		observeSvc,
		cfg.ScreenshotQueueSize,
		cfg.ScreenshotHighQueueSize,
	)
	screenAnalyzer := driver.NewCascadingScreenAnalyzer(cfg, http.DefaultClient, logger)
	if screenAnalyzer.Configured() {
		observationDispatcher.SetScreenAnalyzer(screenAnalyzer)
	}

	grpcServer := grpc.NewServer()
	grpcHandler := handler.NewObserverHandler(observeSvc, screenshotSvc, logger)
	grpcHandler.Register(grpcServer)

	httpHandler := handler.NewHTTPHandler(
		storage,
		observationDispatcher,
		time.Duration(cfg.ScreenshotTimeoutSec)*time.Second,
		time.Duration(cfg.DumpUITimeoutSec)*time.Second,
		logger,
	)
	healthServer := &http.Server{Addr: cfg.HealthAddr, Handler: httpHandler.Routes()}

	go func() {
		listenConfig := net.ListenConfig{}
		lis, err := listenConfig.Listen(ctx, "tcp", cfg.GRPCAddr)
		if err != nil {
			logger.Error("grpc listen", "error", err)
			os.Exit(1)
		}
		logger.Info("grpc server started", "addr", cfg.GRPCAddr)
		_ = grpcServer.Serve(lis)
	}()

	go func() {
		logger.Info("health server started", "addr", cfg.HealthAddr)
		_ = healthServer.ListenAndServe()
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	grpcServer.GracefulStop()
	_ = healthServer.Shutdown(shutdownCtx)
}
