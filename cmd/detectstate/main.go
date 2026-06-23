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

type detectStateRequest struct {
	Serial          string `json:"serial"`
	Mode            string `json:"mode"`
	Platform        string `json:"platform"`
	UseScreenshot   bool   `json:"use_screenshot"`
	StoreScreenshot bool   `json:"store_screenshot"`
	TimeoutSec      int    `json:"timeout_sec"`
	Priority        string `json:"priority"`
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr, http.DefaultClient); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer, client *http.Client) error {
	fs := flag.NewFlagSet("detectstate", flag.ContinueOnError)
	fs.SetOutput(stderr)

	baseURL := fs.String("url", envDefault(getenv, "OBSERVER_HTTP_URL", "http://localhost:9090"), "observer HTTP URL")
	serial := fs.String("serial", envDefault(getenv, "SERIAL", ""), "phone serial")
	mode := fs.String("mode", envDefault(getenv, "DETECT_MODE", "auto"), "detect mode: auto, ui or vlm")
	platform := fs.String("platform", envDefault(getenv, "DETECT_PLATFORM", "android"), "platform hint for VLM")
	useScreenshot := fs.Bool("use-screenshot", envBoolDefault(getenv, "DETECT_USE_SCREENSHOT", true), "capture screenshot for VLM")
	storeScreenshot := fs.Bool("store-screenshot", envBoolDefault(getenv, "DETECT_STORE_SCREENSHOT", false), "store screenshot in MinIO")
	priority := fs.String("priority", envDefault(getenv, "DETECT_PRIORITY", "normal"), "detect priority: normal or high")
	timeoutSec := fs.Int("timeout-sec", envIntDefault(getenv, "DETECT_TIMEOUT_SEC", 30), "request timeout in seconds")
	autoStart := fs.Bool("auto-start", envBoolDefault(getenv, "OBSERVER_AUTO_START", true), "start local observer if it is not running")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*serial) == "" {
		return errors.New("SERIAL обязателен: make detect-state SERIAL=R83YA05Y51P")
	}
	if *timeoutSec <= 0 {
		return errors.New("DETECT_TIMEOUT_SEC должен быть больше 0")
	}
	if strings.TrimSpace(*mode) == "vlm" && !*useScreenshot {
		return errors.New("DETECT_MODE=vlm требует DETECT_USE_SCREENSHOT=true")
	}

	payload, err := json.Marshal(detectStateRequest{
		Serial:          strings.TrimSpace(*serial),
		Mode:            strings.TrimSpace(*mode),
		Platform:        strings.TrimSpace(*platform),
		UseScreenshot:   *useScreenshot,
		StoreScreenshot: *storeScreenshot,
		TimeoutSec:      *timeoutSec,
		Priority:        strings.TrimSpace(*priority),
	})
	if err != nil {
		return err
	}

	httpClient := *client
	httpClient.Timeout = time.Duration(*timeoutSec+5) * time.Second

	endpoint := detectStateEndpoint(*baseURL)
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

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "  "); err != nil {
		return err
	}
	if _, err := pretty.WriteTo(stdout); err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout)
	return err
}

func detectStateEndpoint(raw string) string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if strings.HasSuffix(raw, "/detect-state") {
		return raw
	}
	return raw + "/detect-state"
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
