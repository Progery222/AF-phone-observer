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

type waitForElementRequest struct {
	Serial          string           `json:"serial"`
	Element         waitElementQuery `json:"element"`
	TimeoutSec      int              `json:"timeout_sec"`
	CheckIntervalMS int              `json:"check_interval_ms"`
	Priority        string           `json:"priority"`
}

type waitElementQuery struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	ResourceID  string `json:"resource_id"`
	ContentDesc string `json:"content_desc"`
	Hint        string `json:"hint"`
	Match       string `json:"match"`
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr, http.DefaultClient); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer, client *http.Client) error {
	fs := flag.NewFlagSet("waitforelement", flag.ContinueOnError)
	fs.SetOutput(stderr)

	baseURL := fs.String("url", envDefault(getenv, "OBSERVER_HTTP_URL", "http://localhost:9090"), "observer HTTP URL")
	serial := fs.String("serial", envDefault(getenv, "SERIAL", ""), "phone serial")
	elementType := fs.String("type", envDefault(getenv, "WAIT_TYPE", ""), "element type/class short name")
	text := fs.String("text", envDefault(getenv, "WAIT_TEXT", ""), "element text")
	resourceID := fs.String("resource-id", envDefault(getenv, "WAIT_RESOURCE_ID", ""), "element resource-id")
	contentDesc := fs.String("content-desc", envDefault(getenv, "WAIT_CONTENT_DESC", ""), "element content-desc")
	hint := fs.String("hint", envDefault(getenv, "WAIT_HINT", ""), "element hint")
	match := fs.String("match", envDefault(getenv, "WAIT_MATCH", "exact"), "match mode: exact or contains")
	priority := fs.String("priority", envDefault(getenv, "WAIT_PRIORITY", "normal"), "wait priority: normal or high")
	timeoutSec := fs.Int("timeout-sec", envIntDefault(getenv, "WAIT_TIMEOUT_SEC", 30), "request timeout in seconds")
	checkIntervalMS := fs.Int("check-interval-ms", envIntDefault(getenv, "WAIT_CHECK_INTERVAL_MS", 500), "poll interval in milliseconds")
	autoStart := fs.Bool("auto-start", envBoolDefault(getenv, "OBSERVER_AUTO_START", true), "start local observer if it is not running")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*serial) == "" {
		return errors.New("SERIAL обязателен: make wait-for-element SERIAL=R83YA05Y51P WAIT_TEXT=OK")
	}
	if *timeoutSec <= 0 {
		return errors.New("WAIT_TIMEOUT_SEC должен быть больше 0")
	}
	if *checkIntervalMS < 100 {
		return errors.New("WAIT_CHECK_INTERVAL_MS должен быть не меньше 100")
	}

	query := waitElementQuery{
		Type:        strings.TrimSpace(*elementType),
		Text:        strings.TrimSpace(*text),
		ResourceID:  strings.TrimSpace(*resourceID),
		ContentDesc: strings.TrimSpace(*contentDesc),
		Hint:        strings.TrimSpace(*hint),
		Match:       strings.TrimSpace(*match),
	}
	if query.Type == "" && query.Text == "" && query.ResourceID == "" && query.ContentDesc == "" && query.Hint == "" {
		return errors.New("нужен хотя бы один селектор: WAIT_TEXT, WAIT_RESOURCE_ID, WAIT_CONTENT_DESC, WAIT_HINT или WAIT_TYPE")
	}

	payload, err := json.Marshal(waitForElementRequest{
		Serial:          strings.TrimSpace(*serial),
		Element:         query,
		TimeoutSec:      *timeoutSec,
		CheckIntervalMS: *checkIntervalMS,
		Priority:        strings.TrimSpace(*priority),
	})
	if err != nil {
		return err
	}

	httpClient := *client
	httpClient.Timeout = time.Duration(*timeoutSec+5) * time.Second

	endpoint := waitForElementEndpoint(*baseURL)
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

func waitForElementEndpoint(raw string) string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if strings.HasSuffix(raw, "/wait-for-element") {
		return raw
	}
	return raw + "/wait-for-element"
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
