package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type StartFunc func(context.Context, string, func(string) string, io.Writer) (func(), error)

type AutoStartConfig struct {
	Enabled bool
	Getenv  func(string) string
	Stderr  io.Writer
	Start   StartFunc
}

func DoWithAutoStart(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	newRequest func() (*http.Request, error),
	cfg AutoStartConfig,
) (*http.Response, func(), error) {
	req, err := newRequest()
	if err != nil {
		return nil, nil, err
	}
	resp, err := client.Do(req)
	if err == nil {
		return resp, nil, nil
	}
	if !cfg.Enabled || !shouldAutoStart(endpoint, err) {
		return nil, nil, err
	}

	start := cfg.Start
	if start == nil {
		start = StartLocalObserver
	}
	getenv := cfg.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	stderr := cfg.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	cleanup, startErr := start(ctx, endpoint, getenv, stderr)
	if startErr != nil {
		return nil, nil, fmt.Errorf("%w; не удалось запустить локальный observer: %w", err, startErr)
	}

	req, err = newRequest()
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	resp, err = client.Do(req)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return resp, cleanup, nil
}

func StartLocalObserver(ctx context.Context, endpoint string, getenv func(string) string, stderr io.Writer) (func(), error) {
	healthAddr, healthURL, ok, err := localObserverAddresses(endpoint)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("auto-start доступен только для localhost/127.0.0.1")
	}

	goBin := envDefault(getenv, "GO", "go")
	cmd := exec.CommandContext(ctx, goBin, "run", "./cmd/server")
	cmd.Env = withEnv(os.Environ(),
		"HEALTH_ADDR", healthAddr,
		"GRPC_ADDR", envDefault(getenv, "GRPC_ADDR", "127.0.0.1:0"),
		"LOG_LEVEL", envDefault(getenv, "LOG_LEVEL", "error"),
	)
	cmd.Stdout = stderr
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			select {
			case <-done:
				return
			default:
			}
			if cmd.Process != nil {
				if err := cmd.Process.Signal(os.Interrupt); err != nil {
					_ = cmd.Process.Kill()
				}
			}
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
				<-done
			}
		})
	}

	if err := waitHealth(ctx, healthURL, done); err != nil {
		cleanup()
		return nil, err
	}
	return cleanup, nil
}

func shouldAutoStart(endpoint string, err error) bool {
	_, _, ok, parseErr := localObserverAddresses(endpoint)
	if parseErr != nil || !ok {
		return false
	}
	return isConnectionRefused(err)
}

// isConnectionRefused распознаёт «соединение отклонено» кросс-платформенно:
// на Unix текст ошибки содержит "connection refused", а на Windows —
// "target machine actively refused it" (connectex). errors.Is покрывает
// случаи, когда ошибка оборачивает syscall.ECONNREFUSED.
func isConnectionRefused(err error) bool {
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "actively refused")
}

func localObserverAddresses(endpoint string) (string, string, bool, error) {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return "", "", false, err
	}
	host := u.Hostname()
	if host == "" || !isLocalHost(host) {
		return "", "", false, nil
	}
	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		default:
			return "", "", false, fmt.Errorf("неизвестная схема URL: %s", u.Scheme)
		}
	}
	healthURL := *u
	healthURL.Path = "/health"
	healthURL.RawQuery = ""
	healthURL.Fragment = ""
	return net.JoinHostPort(host, port), healthURL.String(), true, nil
}

func isLocalHost(host string) bool {
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

func waitHealth(ctx context.Context, healthURL string, done <-chan error) error {
	waitCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		req, err := http.NewRequestWithContext(waitCtx, http.MethodGet, healthURL, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case err := <-done:
			return fmt.Errorf("observer завершился до health check: %w", err)
		case <-waitCtx.Done():
			return waitCtx.Err()
		case <-ticker.C:
		}
	}
}

func withEnv(base []string, keyValues ...string) []string {
	env := append([]string(nil), base...)
	for i := 0; i+1 < len(keyValues); i += 2 {
		key := keyValues[i]
		value := keyValues[i+1]
		prefix := key + "="
		replaced := false
		for idx, item := range env {
			if strings.HasPrefix(item, prefix) {
				env[idx] = prefix + value
				replaced = true
				break
			}
		}
		if !replaced {
			env = append(env, prefix+value)
		}
	}
	return env
}

func envDefault(getenv func(string) string, key, fallback string) string {
	if getenv != nil {
		if v := getenv(key); v != "" {
			return v
		}
	}
	return fallback
}
