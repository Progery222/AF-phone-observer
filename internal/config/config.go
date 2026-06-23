package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	GRPCAddr        string
	HealthAddr      string
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucket     string
	MinioUseSSL     bool
	ScreenshotTmpDir string
	LogLevel        slog.Level
}

func Load() Config {
	return Config{
		GRPCAddr:         env("GRPC_ADDR", ":50053"),
		HealthAddr:       env("HEALTH_ADDR", ":9090"),
		MinioEndpoint:    env("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:   env("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:   env("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:      env("MINIO_BUCKET", "af-screenshots"),
		MinioUseSSL:      env("MINIO_USE_SSL", "false") == "true",
		ScreenshotTmpDir: env("SCREENSHOT_TMP_DIR", os.TempDir()),
		LogLevel:         parseLogLevel(env("LOG_LEVEL", "info")),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseLogLevel(raw string) slog.Level {
	switch strings.ToLower(raw) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
