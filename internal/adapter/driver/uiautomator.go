package driver

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"os/exec"
	"strings"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
	"github.com/mobilefarm/af/phone-observer/internal/port"
)

var adbCommandContext = exec.CommandContext

type UIAutomatorDriver struct {
	log port.Logger
}

func NewUIAutomatorDriver(log port.Logger) *UIAutomatorDriver {
	return &UIAutomatorDriver{log: log}
}

func (u *UIAutomatorDriver) DumpUI(ctx context.Context, serial string) (domain.ScreenState, error) {
	if IsTestSerial(serial) {
		return domain.ScreenState{
			Serial: serial,
			XMLDump: `<hierarchy><node text="Войти" content-desc="login" class="android.widget.Button" bounds="[400,500][680,580]" />` +
				`<node text="OK" resource-id="stub:id/ok" content-desc="Create" /></hierarchy>`,
		}, nil
	}
	remote := fmt.Sprintf("/sdcard/af_window_dump_%d.xml", time.Now().UnixNano())
	_ = adbCommandContext(ctx, "adb", "-s", serial, "shell", "rm", "-f", remote).Run()
	dumpOut, err := adbCommandContext(ctx, "adb", "-s", serial, "shell", "uiautomator", "dump", remote).CombinedOutput()
	if err != nil || dumpHasError(dumpOut) {
		return domain.ScreenState{}, fmt.Errorf("%w: uiautomator dump: %s", domain.ErrUIDumpFailed, strings.TrimSpace(string(dumpOut)))
	}
	out, err := adbCommandContext(ctx, "adb", "-s", serial, "shell", "cat", remote).Output()
	if err != nil {
		return domain.ScreenState{}, err
	}
	xmlDump := string(out)
	if !strings.Contains(xmlDump, "<hierarchy") {
		return domain.ScreenState{}, fmt.Errorf("%w: uiautomator dump не вернул hierarchy", domain.ErrUIDumpFailed)
	}
	return domain.ScreenState{Serial: serial, XMLDump: xmlDump}, nil
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
	return adbCommandContext(ctx, "adb", "version").Run()
}

var _ port.UIDumper = (*UIAutomatorDriver)(nil)

type AdbScreenshotDriver struct {
	log port.Logger
}

func NewAdbScreenshotDriver(log port.Logger) *AdbScreenshotDriver {
	return &AdbScreenshotDriver{log: log}
}

func (a *AdbScreenshotDriver) Capture(ctx context.Context, serial string) (domain.Screenshot, error) {
	takenAt := time.Now().UTC()
	var out []byte
	var err error
	if IsTestSerial(serial) {
		out, err = stubScreenshotPNG()
	} else {
		out, err = adbCommandContext(ctx, "adb", "-s", serial, "exec-out", "screencap", "-p").Output()
	}
	if err != nil {
		return domain.Screenshot{}, err
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		return domain.Screenshot{}, err
	}
	return domain.Screenshot{
		Serial:    serial,
		ObjectKey: takenAt.Format("20060102-150405"),
		Bytes:     out,
		SizeBytes: int64(len(out)),
		Width:     int32(cfg.Width),
		Height:    int32(cfg.Height),
		TakenAt:   takenAt,
	}, nil
}

func stubScreenshotPNG() ([]byte, error) {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ port.ScreenshotCapture = (*AdbScreenshotDriver)(nil)

func dumpHasError(out []byte) bool {
	lowered := strings.ToLower(string(out))
	return strings.Contains(lowered, "error:") || strings.Contains(lowered, "could not get idle state")
}
