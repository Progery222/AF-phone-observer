package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
	"github.com/mobilefarm/af/phone-observer/internal/service"
)

type stubDumper struct{}

func (stubDumper) DumpUI(_ context.Context, serial string) (domain.ScreenState, error) {
	return domain.ScreenState{Serial: serial, XMLDump: "<hierarchy/>"}, nil
}
func (stubDumper) DetectState(_ context.Context, serial string) (domain.ScreenState, error) {
	return domain.ScreenState{Serial: serial, PackageName: "com.example"}, nil
}
func (stubDumper) Ping(context.Context) error { return nil }

func TestObserveService_DumpUI(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := service.NewObserveService(stubDumper{}, log)
	state, err := svc.DumpUI(context.Background(), "stub")
	if err != nil {
		t.Fatal(err)
	}
	if state.XMLDump == "" {
		t.Fatal("expected xml dump")
	}
}
