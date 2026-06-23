package driver

import (
	"bytes"
	"context"
	"image/png"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestUIAutomatorDriverDoesNotReadStaleDumpWhenDumpFails(t *testing.T) {
	restore := stubADBCommand("dump_fail_cat_stale")
	defer restore()

	driver := NewUIAutomatorDriver(slog.New(slog.DiscardHandler))
	state, err := driver.DumpUI(context.Background(), "R5GL2218DMR")
	if err == nil {
		t.Fatalf("expected dump error, got state: %+v", state)
	}
	if strings.Contains(state.XMLDump, "STALE") {
		t.Fatalf("must not return stale XML: %+v", state)
	}
	if !strings.Contains(err.Error(), "uiautomator dump") {
		t.Fatalf("expected dump error, got %v", err)
	}
}

func TestUIAutomatorDriverReadsFreshDumpAfterSuccessfulDump(t *testing.T) {
	restore := stubADBCommand("dump_success")
	defer restore()

	driver := NewUIAutomatorDriver(slog.New(slog.DiscardHandler))
	state, err := driver.DumpUI(context.Background(), "R5GL2218DMR")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(state.XMLDump, "FRESH") {
		t.Fatalf("expected fresh XML, got %q", state.XMLDump)
	}
}

func TestAdbScreenshotDriverStubReturnsValidPNGMetadata(t *testing.T) {
	driver := NewAdbScreenshotDriver(slog.New(slog.DiscardHandler))

	shot, err := driver.Capture(context.Background(), "stub")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := png.Decode(bytes.NewReader(shot.Bytes)); err != nil {
		t.Fatalf("stub screenshot must be a valid png: %v", err)
	}
	if shot.SizeBytes != int64(len(shot.Bytes)) {
		t.Fatalf("expected size %d, got %d", len(shot.Bytes), shot.SizeBytes)
	}
	if shot.Width != 1 || shot.Height != 1 {
		t.Fatalf("expected 1x1 png, got %dx%d", shot.Width, shot.Height)
	}
	if shot.TakenAt.IsZero() {
		t.Fatal("expected taken_at to be set")
	}
}

func stubADBCommand(mode string) func() {
	old := adbCommandContext
	adbCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		testArgs := append([]string{"-test.run=TestADBHelperProcess", "--"}, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], testArgs...)
		cmd.Env = append(os.Environ(), "GO_WANT_ADB_HELPER_PROCESS=1", "ADB_HELPER_MODE="+mode)
		return cmd
	}
	return func() {
		adbCommandContext = old
	}
}

func TestADBHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_ADB_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	separator := 0
	for i, arg := range args {
		if arg == "--" {
			separator = i + 1
			break
		}
	}
	if separator == 0 {
		os.Exit(2)
	}
	adbArgs := args[separator:]
	mode := os.Getenv("ADB_HELPER_MODE")
	switch {
	case containsAll(adbArgs, "shell", "uiautomator", "dump"):
		if mode == "dump_fail_cat_stale" {
			_, _ = os.Stdout.WriteString("ERROR: could not get idle state.")
			os.Exit(1)
		}
		_, _ = os.Stdout.WriteString("UI hierchary dumped to: /sdcard/af_window_dump.xml")
		os.Exit(0)
	case containsAll(adbArgs, "shell", "cat"):
		if mode == "dump_fail_cat_stale" {
			_, _ = os.Stdout.WriteString(`<hierarchy><node text="STALE" class="android.widget.TextView" bounds="[0,0][1,1]" /></hierarchy>`)
			os.Exit(0)
		}
		_, _ = os.Stdout.WriteString(`<hierarchy><node text="FRESH" class="android.widget.TextView" bounds="[0,0][1,1]" /></hierarchy>`)
		os.Exit(0)
	case containsAll(adbArgs, "shell", "rm"):
		os.Exit(0)
	default:
		os.Exit(0)
	}
}

func containsAll(args []string, values ...string) bool {
	joined := strings.Join(args, " ")
	for _, value := range values {
		if !strings.Contains(joined, value) {
			return false
		}
	}
	return true
}
