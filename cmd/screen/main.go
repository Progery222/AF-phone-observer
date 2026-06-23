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
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	observercli "github.com/mobilefarm/af/phone-observer/internal/cli"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr, http.DefaultClient); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer, client *http.Client) error {
	fs := flag.NewFlagSet("screen", flag.ContinueOnError)
	fs.SetOutput(stderr)

	baseURL := fs.String("url", envDefault(getenv, "OBSERVER_HTTP_URL", "http://localhost:9090"), "observer HTTP URL")
	serial := fs.String("serial", envDefault(getenv, "SERIAL", ""), "phone serial")
	priority := fs.String("priority", envDefault(getenv, "SCREEN_PRIORITY", "normal"), "priority: normal or high")
	timeoutSec := fs.Int("timeout-sec", envIntDefault(getenv, "SCREEN_TIMEOUT_SEC", 10), "request timeout in seconds")
	autoStart := fs.Bool("auto-start", envBoolDefault(getenv, "OBSERVER_AUTO_START", true), "start local observer if it is not running")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*serial) == "" {
		return errors.New("SERIAL обязателен: make screen SERIAL=R83YA05Y51P")
	}
	if *timeoutSec <= 0 {
		return errors.New("SCREEN_TIMEOUT_SEC должен быть больше 0")
	}

	httpClient := *client
	httpClient.Timeout = time.Duration(*timeoutSec+5) * time.Second

	endpoint := withQuery(screenEndpoint(*baseURL, *serial), map[string]string{
		"priority":    strings.TrimSpace(*priority),
		"timeout_sec": strconv.Itoa(*timeoutSec),
	})
	newRequest := func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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
		return fmt.Errorf("GET %s вернул %s: %s", endpoint, resp.Status, strings.TrimSpace(string(body)))
	}
	return writePrettyJSON(stdout, body)
}

func screenEndpoint(raw, serial string) string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if strings.Contains(raw, "/screen/") {
		return raw
	}
	return raw + "/screen/" + url.PathEscape(strings.TrimSpace(serial))
}

func withQuery(endpoint string, values map[string]string) string {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	query := parsed.Query()
	for key, value := range values {
		if value != "" {
			query.Set(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func writePrettyJSON(stdout io.Writer, body []byte) error {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "  "); err != nil {
		return err
	}
	if _, err := pretty.WriteTo(stdout); err != nil {
		return err
	}
	_, err := fmt.Fprintln(stdout)
	return err
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
