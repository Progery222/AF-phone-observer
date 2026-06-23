package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
	"github.com/mobilefarm/af/phone-observer/internal/port"
)

type ObserveService struct {
	dumper port.UIDumper
	log    port.Logger
}

func NewObserveService(dumper port.UIDumper, log port.Logger) *ObserveService {
	return &ObserveService{dumper: dumper, log: log}
}

func (s *ObserveService) DumpUI(ctx context.Context, serial string) (domain.ScreenState, error) {
	if serial == "" {
		return domain.ScreenState{}, domain.ErrInvalidSerial
	}
	state, err := s.dumper.DumpUI(ctx, serial)
	if err != nil {
		return domain.ScreenState{}, domain.ErrUIDumpFailed
	}
	s.log.Info("ui dump captured", "service", "phone-observer", "serial", serial)
	return state, nil
}

func (s *ObserveService) DumpUIDocument(ctx context.Context, serial string) (domain.UIDump, error) {
	if serial == "" {
		return domain.UIDump{}, domain.ErrInvalidSerial
	}
	state, err := s.dumper.DumpUI(ctx, serial)
	if err != nil {
		return domain.UIDump{}, fmt.Errorf("%w: %w", domain.ErrUIDumpFailed, err)
	}
	dump, err := domain.ParseUIDump(serial, state.XMLDump, time.Now().UTC())
	if err != nil {
		return domain.UIDump{}, fmt.Errorf("%w: %w", domain.ErrUIDumpFailed, err)
	}
	s.log.Info("ui dump captured", "service", "phone-observer", "serial", serial)
	return dump, nil
}

func (s *ObserveService) DetectState(ctx context.Context, serial string) (domain.ScreenState, error) {
	if serial == "" {
		return domain.ScreenState{}, domain.ErrInvalidSerial
	}
	return s.dumper.DetectState(ctx, serial)
}

type ScreenshotService struct {
	capture port.ScreenshotCapture
	storage port.ObjectStorage
	log     port.Logger
}

func NewScreenshotService(capture port.ScreenshotCapture, storage port.ObjectStorage, log port.Logger) *ScreenshotService {
	return &ScreenshotService{capture: capture, storage: storage, log: log}
}

func (s *ScreenshotService) CaptureAndStore(ctx context.Context, serial string) (domain.Screenshot, error) {
	shot, err := s.Capture(ctx, serial)
	if err != nil {
		return domain.Screenshot{}, err
	}
	key := screenshotStorageKey(serial, shot.ObjectKey, shot.TakenAt)
	url, err := s.storage.Upload(ctx, key, shot.Bytes)
	if err != nil {
		return domain.Screenshot{}, fmt.Errorf("%w: %w", domain.ErrStorageUnavailable, err)
	}
	shot.ObjectKey = key
	shot.StorageURL = url
	s.log.Info("screenshot stored", "service", "phone-observer", "serial", serial, "key", key)
	return shot, nil
}

func (s *ScreenshotService) Capture(ctx context.Context, serial string) (domain.Screenshot, error) {
	if serial == "" {
		return domain.Screenshot{}, domain.ErrInvalidSerial
	}
	shot, err := s.capture.Capture(ctx, serial)
	if err != nil {
		return domain.Screenshot{}, fmt.Errorf("%w: %w", domain.ErrScreenshotFailed, err)
	}
	if shot.TakenAt.IsZero() {
		shot.TakenAt = time.Now().UTC()
	}
	if shot.SizeBytes == 0 {
		shot.SizeBytes = int64(len(shot.Bytes))
	}
	return shot, nil
}

func screenshotStorageKey(serial, objectName string, takenAt time.Time) string {
	if objectName == "" {
		objectName = takenAt.Format("20060102-150405")
	}
	objectName = strings.TrimPrefix(objectName, serial+"/")
	objectName = strings.TrimSuffix(objectName, ".png")
	return fmt.Sprintf("%s/%s.png", serial, objectName)
}
