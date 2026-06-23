package cli

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDoWithAutoStartRetriesLocalConnectionRefused(t *testing.T) {
	attempts := 0
	started := false
	stopped := false
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("Post \"http://127.0.0.1:19093/dump-ui\": dial tcp 127.0.0.1:19093: connect: connection refused")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, nil
	})}

	resp, cleanup, err := DoWithAutoStart(
		context.Background(),
		client,
		"http://127.0.0.1:19093/dump-ui",
		func() (*http.Request, error) {
			return http.NewRequestWithContext(context.Background(), http.MethodPost, "http://127.0.0.1:19093/dump-ui", strings.NewReader(`{}`))
		},
		AutoStartConfig{
			Enabled: true,
			Start: func(context.Context, string, func(string) string, io.Writer) (func(), error) {
				started = true
				return func() { stopped = true }, nil
			},
		},
	)
	if err != nil {
		t.Fatalf("DoWithAutoStart returned error: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if attempts != 2 {
		t.Fatalf("expected two attempts, got %d", attempts)
	}
	if !started {
		t.Fatal("expected local observer start")
	}
	if cleanup == nil {
		t.Fatal("expected cleanup")
	}
	cleanup()
	if !stopped {
		t.Fatal("expected local observer cleanup")
	}
}

func TestDoWithAutoStartDoesNotStartForRemoteURL(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("Post \"http://example.com/dump-ui\": dial tcp 203.0.113.10:9090: connect: connection refused")
	})}

	resp, cleanup, err := DoWithAutoStart(
		context.Background(),
		client,
		"http://example.com/dump-ui",
		func() (*http.Request, error) {
			return http.NewRequestWithContext(context.Background(), http.MethodPost, "http://example.com/dump-ui", strings.NewReader(`{}`))
		},
		AutoStartConfig{
			Enabled: true,
			Start: func(context.Context, string, func(string) string, io.Writer) (func(), error) {
				t.Fatal("remote URL must not start local observer")
				return nil, nil
			},
		},
	)
	if err == nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		t.Fatal("expected original request error")
	}
	if resp != nil {
		_ = resp.Body.Close()
	}
	if cleanup != nil {
		t.Fatal("unexpected cleanup")
	}
}

func TestLocalObserverAddressesFromEndpoint(t *testing.T) {
	healthAddr, healthURL, ok, err := localObserverAddresses("http://localhost:19093/dump-ui")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected local endpoint")
	}
	if healthAddr != "localhost:19093" {
		t.Fatalf("unexpected health addr: %s", healthAddr)
	}
	if healthURL != "http://localhost:19093/health" {
		t.Fatalf("unexpected health url: %s", healthURL)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
