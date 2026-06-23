package port

import (
	"context"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
)

type UIDumper interface {
	DumpUI(ctx context.Context, serial string) (domain.ScreenState, error)
	DetectState(ctx context.Context, serial string) (domain.ScreenState, error)
	Ping(ctx context.Context) error
}

type ScreenshotCapture interface {
	Capture(ctx context.Context, serial string) (domain.Screenshot, error)
}

type ScreenAnalyzer interface {
	Analyze(ctx context.Context, serial, platform string, screenshot []byte) (domain.VLMAnalysis, error)
}

type ObjectStorage interface {
	Upload(ctx context.Context, key string, data []byte) (string, error)
	Ping(ctx context.Context) error
}
