package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	GRPCAddr                string
	HealthAddr              string
	MinioEndpoint           string
	MinioAccessKey          string
	MinioSecretKey          string
	MinioBucket             string
	MinioUseSSL             bool
	MinioPublicBaseURL      string
	ScreenshotTmpDir        string
	ScreenshotTimeoutSec    int
	DumpUITimeoutSec        int
	ScreenshotQueueSize     int
	ScreenshotHighQueueSize int
	VLMBackends             string
	VisionServerURL         string
	OllamaURL               string
	OllamaVLMModel          string
	OpenAIAPIKey            string
	OpenAIBaseURL           string
	OpenAIModel             string
	VLMTimeoutSec           int
	VLMMaxConcurrency       int
	LogLevel                slog.Level
}

func Load() Config {
	return Config{
		GRPCAddr:                env("GRPC_ADDR", ":50053"),
		HealthAddr:              env("HEALTH_ADDR", ":9090"),
		MinioEndpoint:           env("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:          env("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:          env("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:             env("MINIO_BUCKET", "af-screenshots"),
		MinioUseSSL:             env("MINIO_USE_SSL", "false") == "true",
		MinioPublicBaseURL:      env("MINIO_PUBLIC_BASE_URL", "http://127.0.0.1:9000"),
		ScreenshotTmpDir:        env("SCREENSHOT_TMP_DIR", os.TempDir()),
		ScreenshotTimeoutSec:    envInt("SCREENSHOT_TIMEOUT_SEC", 10),
		DumpUITimeoutSec:        envInt("DUMP_UI_TIMEOUT_SEC", 30),
		ScreenshotQueueSize:     envInt("SCREENSHOT_QUEUE_SIZE", 32),
		ScreenshotHighQueueSize: envInt("SCREENSHOT_HIGH_QUEUE_SIZE", 8),
		VLMBackends:             env("VLM_BACKENDS", ""),
		VisionServerURL:         env("VISION_SERVER_URL", ""),
		OllamaURL:               env("OLLAMA_URL", "http://localhost:11434"),
		OllamaVLMModel:          env("OLLAMA_VLM_MODEL", "qwen2.5vl:7b"),
		OpenAIAPIKey:            env("OPENAI_API_KEY", ""),
		OpenAIBaseURL:           env("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:             env("OPENAI_MODEL", "gpt-5.4-mini"),
		VLMTimeoutSec:           envInt("VLM_TIMEOUT_SEC", 20),
		VLMMaxConcurrency:       envInt("VLM_MAX_CONCURRENCY", 2),
		LogLevel:                parseLogLevel(env("LOG_LEVEL", "info")),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := env(key, "")
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
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
