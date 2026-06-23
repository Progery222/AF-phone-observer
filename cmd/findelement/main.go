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

type findElementRequest struct {
	Serial     string           `json:"serial"`
	Element    findElementQuery `json:"element"`
	TimeoutSec int              `json:"timeout_sec"`
	Priority   string           `json:"priority"`
}

type findElementQuery struct {
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
	fs := flag.NewFlagSet("findelement", flag.ContinueOnError)
	fs.SetOutput(stderr)

	baseURL := fs.String("url", envDefault(getenv, "OBSERVER_HTTP_URL", "http://localhost:9090"), "observer HTTP URL")
	serial := fs.String("serial", envDefault(getenv, "SERIAL", ""), "phone serial")
	elementType := fs.String("type", envDefault(getenv, "FIND_TYPE", ""), "element type/class short name")
	text := fs.String("text", envDefault(getenv, "FIND_TEXT", ""), "element text")
	resourceID := fs.String("resource-id", envDefault(getenv, "FIND_RESOURCE_ID", ""), "element resource-id")
	contentDesc := fs.String("content-desc", envDefault(getenv, "FIND_CONTENT_DESC", ""), "element content-desc")
	hint := fs.String("hint", envDefault(getenv, "FIND_HINT", ""), "element hint")
	match := fs.String("match", envDefault(getenv, "FIND_MATCH", "exact"), "match mode: exact or contains")
	priority := fs.String("priority", envDefault(getenv, "FIND_PRIORITY", "normal"), "find priority: normal or high")
	timeoutSec := fs.Int("timeout-sec", envIntDefault(getenv, "FIND_TIMEOUT_SEC", 30), "request timeout in seconds")
	autoStart := fs.Bool("auto-start", envBoolDefault(getenv, "OBSERVER_AUTO_START", true), "start local observer if it is not running")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*serial) == "" {
		return errors.New("SERIAL обязателен: make find-element SERIAL=R83YA05Y51P FIND_TEXT=OK")
	}
	if *timeoutSec <= 0 {
		return errors.New("FIND_TIMEOUT_SEC должен быть больше 0")
	}

	query := findElementQuery{
		Type:        strings.TrimSpace(*elementType),
		Text:        strings.TrimSpace(*text),
		ResourceID:  strings.TrimSpace(*resourceID),
		ContentDesc: strings.TrimSpace(*contentDesc),
		Hint:        strings.TrimSpace(*hint),
		Match:       strings.TrimSpace(*match),
	}
	if query.Type == "" && query.Text == "" && query.ResourceID == "" && query.ContentDesc == "" && query.Hint == "" {
		return errors.New("нужен хотя бы один селектор: FIND_TEXT, FIND_RESOURCE_ID, FIND_CONTENT_DESC, FIND_HINT или FIND_TYPE")
	}

	payload, err := json.Marshal(findElementRequest{
		Serial:     strings.TrimSpace(*serial),
		Element:    query,
		TimeoutSec: *timeoutSec,
		Priority:   strings.TrimSpace(*priority),
	})
	if err != nil {
		return err
	}

	httpClient := *client
	httpClient.Timeout = time.Duration(*timeoutSec+5) * time.Second

	endpoint := findElementEndpoint(*baseURL)
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

func findElementEndpoint(raw string) string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if strings.HasSuffix(raw, "/find-element") {
		return raw
	}
	return raw + "/find-element"
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
