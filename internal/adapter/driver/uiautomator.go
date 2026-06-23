package driver

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
	"github.com/mobilefarm/af/phone-observer/internal/port"
)

type UIAutomatorDriver struct {
	log port.Logger
}

func NewUIAutomatorDriver(log port.Logger) *UIAutomatorDriver {
	return &UIAutomatorDriver{log: log}
}

func (u *UIAutomatorDriver) DumpUI(ctx context.Context, serial string) (domain.ScreenState, error) {
	if serial == "stub" {
		return domain.ScreenState{Serial: serial, XMLDump: "<hierarchy/>"}, nil
	}
	remote := "/sdcard/window_dump.xml"
	_, _ = exec.CommandContext(ctx, "adb", "-s", serial, "shell", "uiautomator", "dump", remote).CombinedOutput()
	out, err := exec.CommandContext(ctx, "adb", "-s", serial, "shell", "cat", remote).Output()
	if err != nil {
		return domain.ScreenState{}, err
	}
	return domain.ScreenState{Serial: serial, XMLDump: string(out)}, nil
}

func (u *UIAutomatorDriver) DetectState(ctx context.Context, serial string) (domain.ScreenState, error) {
	state, err := u.DumpUI(ctx, serial)
	if err != nil {
		return state, err
	}
	if strings.Contains(state.XMLDump, "package=") {
		start := strings.Index(state.XMLDump, `package="`)
		if start >= 0 {
			start += len(`package="`)
			end := strings.Index(state.XMLDump[start:], `"`)
			if end > 0 {
				state.PackageName = state.XMLDump[start : start+end]
			}
		}
	}
	return state, nil
}

func (u *UIAutomatorDriver) Ping(ctx context.Context) error {
	return exec.CommandContext(ctx, "adb", "version").Run()
}

var _ port.UIDumper = (*UIAutomatorDriver)(nil)

type AdbScreenshotDriver struct {
	log port.Logger
}

func NewAdbScreenshotDriver(log port.Logger) *AdbScreenshotDriver {
	return &AdbScreenshotDriver{log: log}
}

func (a *AdbScreenshotDriver) Capture(ctx context.Context, serial string) (domain.Screenshot, error) {
	if serial == "stub" {
		return domain.Screenshot{Serial: serial, ObjectKey: time.Now().Format("20060102-150405"), Bytes: []byte("stub")}, nil
	}
	out, err := exec.CommandContext(ctx, "adb", "-s", serial, "exec-out", "screencap", "-p").Output()
	if err != nil {
		return domain.Screenshot{}, err
	}
	return domain.Screenshot{
		Serial:    serial,
		ObjectKey: time.Now().Format("20060102-150405"),
		Bytes:     out,
	}, nil
}

var _ port.ScreenshotCapture = (*AdbScreenshotDriver)(nil)
