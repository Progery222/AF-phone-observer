package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunPostsWaitForElementRequest(t *testing.T) {
	var got waitForElementRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/wait-for-element" {
			t.Fatalf("expected /wait-for-element, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"found":true,
			"element":{"type":"Button","text":"OK","resource_id":"stub:id/ok","content_desc":"Create","hint":"","bounds":{"x1":0,"y1":0,"x2":2,"y2":2},"center":{"x":1,"y":1}},
			"found_by":"text",
			"wait_time_ms":500,
			"check_count":2
		}`))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(
		context.Background(),
		[]string{"-url=" + server.URL, "-serial=stub", "-text=OK", "-content-desc=Create", "-match=contains", "-priority=high", "-timeout-sec=30", "-check-interval-ms=500"},
		func(string) string { return "" },
		&stdout,
		&stderr,
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run failed: %v, stderr=%s", err, stderr.String())
	}
	if got.Serial != "stub" || got.Element.Text != "OK" || got.Element.ContentDesc != "Create" || got.Element.Match != "contains" || got.Priority != "high" || got.TimeoutSec != 30 || got.CheckIntervalMS != 500 {
		t.Fatalf("unexpected request payload: %+v", got)
	}
	if !strings.Contains(stdout.String(), `"found": true`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestRunRequiresSerialAndSelector(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), nil, func(string) string { return "" }, &stdout, &stderr, http.DefaultClient)
	if err == nil || !strings.Contains(err.Error(), "SERIAL обязателен") {
		t.Fatalf("unexpected error: %v", err)
	}

	err = run(context.Background(), []string{"-serial=stub"}, func(string) string { return "" }, &stdout, &stderr, http.DefaultClient)
	if err == nil || !strings.Contains(err.Error(), "нужен хотя бы один селектор") {
		t.Fatalf("unexpected selector error: %v", err)
	}
}

func TestWaitForElementEndpointDoesNotDuplicatePath(t *testing.T) {
	got := waitForElementEndpoint("http://localhost:9090/wait-for-element")
	if got != "http://localhost:9090/wait-for-element" {
		t.Fatalf("unexpected endpoint: %s", got)
	}
}
