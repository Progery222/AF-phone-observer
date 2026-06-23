package service

import (
	"context"
	"fmt"

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
	if serial == "" {
		return domain.Screenshot{}, domain.ErrInvalidSerial
	}
	shot, err := s.capture.Capture(ctx, serial)
	if err != nil {
		return domain.Screenshot{}, domain.ErrScreenshotFailed
	}
	key := fmt.Sprintf("%s/%s.png", serial, shot.ObjectKey)
	url, err := s.storage.Upload(ctx, key, shot.Bytes)
	if err != nil {
		return domain.Screenshot{}, domain.ErrStorageUnavailable
	}
	shot.ObjectKey = url
	s.log.Info("screenshot stored", "service", "phone-observer", "serial", serial, "key", key)
	return shot, nil
}
