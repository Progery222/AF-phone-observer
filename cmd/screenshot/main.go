package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	observercli "github.com/mobilefarm/af/phone-observer/internal/cli"
)

type screenshotRequest struct {
	Serial       string `json:"serial"`
	StoreInMinio bool   `json:"store_in_minio"`
	TimeoutSec   int    `json:"timeout_sec"`
	Priority     string `json:"priority"`
}

type screenshotResponse struct {
	ScreenshotURL string `json:"screenshot_url"`
	MinioKey      string `json:"minio_key"`
	SizeBytes     int64  `json:"size_bytes"`
	Resolution    struct {
		Width  int32 `json:"width"`
		Height int32 `json:"height"`
	} `json:"resolution"`
	TakenAt string `json:"taken_at"`
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr, http.DefaultClient); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer, client *http.Client) error {
	fs := flag.NewFlagSet("screenshot", flag.ContinueOnError)
	fs.SetOutput(stderr)

	baseURL := fs.String("url", envDefault(getenv, "OBSERVER_HTTP_URL", "http://localhost:9090"), "observer HTTP URL")
	serial := fs.String("serial", envDefault(getenv, "SERIAL", ""), "phone serial")
	priority := fs.String("priority", envDefault(getenv, "SCREENSHOT_PRIORITY", "normal"), "screenshot priority: normal or high")
	timeoutSec := fs.Int("timeout-sec", envIntDefault(getenv, "SCREENSHOT_TIMEOUT_SEC", 10), "request timeout in seconds")
	storeInMinio := fs.Bool("store-in-minio", envBoolDefault(getenv, "SCREENSHOT_STORE_IN_MINIO", true), "store screenshot in MinIO")
	autoStart := fs.Bool("auto-start", envBoolDefault(getenv, "OBSERVER_AUTO_START", true), "start local observer if it is not running")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*serial) == "" {
		return errors.New("SERIAL обязателен: make screenshot SERIAL=R83YA05Y51P")
	}
	if *timeoutSec <= 0 {
		return errors.New("SCREENSHOT_TIMEOUT_SEC должен быть больше 0")
	}

	payload, err := json.Marshal(screenshotRequest{
		Serial:       strings.TrimSpace(*serial),
		StoreInMinio: *storeInMinio,
		TimeoutSec:   *timeoutSec,
		Priority:     strings.TrimSpace(*priority),
	})
	if err != nil {
		return err
	}

	httpClient := *client
	httpClient.Timeout = time.Duration(*timeoutSec+5) * time.Second

	endpoint := screenshotEndpoint(*baseURL)
	newRequest := func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}
	resp, cleanup, err := observercli.DoWithAutoStart(ctx, &httpClient, endpoint, newRequest, observercli.AutoStartConfig{
		Enabled: *autoStart,
		Getenv:  getenv,
		Stderr:  stderr,
	})
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("POST %s вернул %s: %s", endpoint, resp.Status, strings.TrimSpace(string(body)))
	}

	var result screenshotResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}

func screenshotEndpoint(raw string) string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if strings.HasSuffix(raw, "/screenshot") {
		return raw
	}
	return raw + "/screenshot"
}

func envDefault(getenv func(string) string, key, fallback string) string {
	if v := getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntDefault(getenv func(string) string, key string, fallback int) int {
	v := getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBoolDefault(getenv func(string) string, key string, fallback bool) bool {
	v := getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}
