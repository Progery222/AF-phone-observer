package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
)

type blockingObservationRunner struct {
	started chan string
	release chan error
}

func newBlockingObservationRunner() *blockingObservationRunner {
	return &blockingObservationRunner{
		started: make(chan string, 16),
		release: make(chan error, 16),
	}
}

func (b *blockingObservationRunner) CaptureAndStore(ctx context.Context, serial string) (domain.Screenshot, error) {
	b.started <- "screenshot:" + serial
	select {
	case err := <-b.release:
		return domain.Screenshot{Serial: serial, TakenAt: time.Now().UTC()}, err
	case <-ctx.Done():
		return domain.Screenshot{}, ctx.Err()
	}
}

func (b *blockingObservationRunner) Capture(ctx context.Context, serial string) (domain.Screenshot, error) {
	b.started <- "screenshot-raw:" + serial
	select {
	case err := <-b.release:
		return domain.Screenshot{Serial: serial, Bytes: []byte("png"), TakenAt: time.Now().UTC()}, err
	case <-ctx.Done():
		return domain.Screenshot{}, ctx.Err()
	}
}

func (b *blockingObservationRunner) DumpUIDocument(ctx context.Context, serial string) (domain.UIDump, error) {
	b.started <- "dump-ui:" + serial
	select {
	case err := <-b.release:
		return domain.UIDump{
			Serial:  serial,
			TakenAt: time.Now().UTC(),
			Elements: []domain.UIElement{
				{Text: "OK", ResourceID: "stub:id/ok"},
			},
			ElementCount: 1,
		}, err
	case <-ctx.Done():
		return domain.UIDump{}, ctx.Err()
	}
}

func TestObservationDispatcherSerializesScreenshotAndDumpUIForSameSerial(t *testing.T) {
	runner := newBlockingObservationRunner()
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	firstDone := runObservationCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, runner.started, "screenshot:same")

	secondDone := runObservationDumpUI(ctx, dispatcher, "same", PriorityNormal)
	assertNoStart(t, runner.started)

	runner.release <- nil
	assertNoError(t, <-firstDone)
	waitStarted(t, runner.started, "dump-ui:same")

	runner.release <- nil
	assertNoError(t, <-secondDone)
}

func TestObservationDispatcherRunsDifferentSerialsInParallel(t *testing.T) {
	runner := newBlockingObservationRunner()
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	firstDone := runObservationDumpUI(ctx, dispatcher, "a", PriorityNormal)
	waitStarted(t, runner.started, "dump-ui:a")

	secondDone := runObservationCapture(ctx, dispatcher, "b", PriorityNormal)
	waitStarted(t, runner.started, "screenshot:b")

	runner.release <- nil
	runner.release <- nil
	assertNoError(t, <-firstDone)
	assertNoError(t, <-secondDone)
}

func TestObservationDispatcherHighPriorityOvertakesPendingNormalAcrossTaskTypes(t *testing.T) {
	runner := newBlockingObservationRunner()
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	running := runObservationCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, runner.started, "screenshot:same")

	pendingNormal := runObservationDumpUI(ctx, dispatcher, "same", PriorityNormal)
	pendingHigh := runObservationDumpUI(ctx, dispatcher, "same", PriorityHigh)
	waitObservationQueueLengths(t, dispatcher, "same", 1, 1)

	runner.release <- nil
	assertNoError(t, <-running)
	waitStarted(t, runner.started, "dump-ui:same")

	runner.release <- nil
	assertNoError(t, <-pendingHigh)
	assertNotDone(t, pendingNormal)

	waitStarted(t, runner.started, "dump-ui:same")
	runner.release <- nil
	assertNoError(t, <-pendingNormal)
}

func TestObservationDispatcherHighPriorityDoesNotInterruptRunningTask(t *testing.T) {
	runner := newBlockingObservationRunner()
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	running := runObservationDumpUI(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, runner.started, "dump-ui:same")

	pendingHigh := runObservationCapture(ctx, dispatcher, "same", PriorityHigh)
	assertNoStart(t, runner.started)

	runner.release <- nil
	assertNoError(t, <-running)
	waitStarted(t, runner.started, "screenshot:same")

	runner.release <- nil
	assertNoError(t, <-pendingHigh)
}

func TestObservationDispatcherSerializesFindElementWithOtherTasks(t *testing.T) {
	runner := newBlockingObservationRunner()
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	running := runObservationCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, runner.started, "screenshot:same")

	pendingFind := runObservationFindElement(ctx, dispatcher, "same", domain.FindElementQuery{Text: "OK"}, PriorityNormal)
	assertNoStart(t, runner.started)

	runner.release <- nil
	assertNoError(t, <-running)
	waitStarted(t, runner.started, "dump-ui:same")

	runner.release <- nil
	assertNoError(t, <-pendingFind)
}

func TestObservationDispatcherWaitForElementFoundAfterSeveralPolls(t *testing.T) {
	runner := &sequenceObservationRunner{
		dumps: []domain.UIDump{
			{Elements: []domain.UIElement{{Text: "Loading"}}},
			{Elements: []domain.UIElement{{Text: "Still loading"}}},
			{Elements: []domain.UIElement{{Text: "OK", ResourceID: "stub:id/ok"}}},
		},
	}
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	result, err := dispatcher.WaitForElement(context.Background(), "stub", domain.FindElementQuery{Text: "OK"}, PriorityNormal, 100*time.Millisecond, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Found || result.FoundBy != "text" || result.CheckCount != 3 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestObservationDispatcherWaitForElementTimeout(t *testing.T) {
	runner := &sequenceObservationRunner{
		dumps: []domain.UIDump{
			{Elements: []domain.UIElement{{Text: "Loading"}}},
		},
	}
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	result, err := dispatcher.WaitForElement(context.Background(), "stub", domain.FindElementQuery{Text: "Missing"}, PriorityNormal, 5*time.Millisecond, time.Millisecond)
	if !errors.Is(err, domain.ErrElementWaitTimeout) {
		t.Fatalf("expected ErrElementWaitTimeout, got %v", err)
	}
	if result.Found || result.CheckCount == 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestObservationDispatcherSerializesWaitForElementWithOtherTasks(t *testing.T) {
	runner := newBlockingObservationRunner()
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	running := runObservationCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, runner.started, "screenshot:same")

	pendingWait := runObservationWaitForElement(ctx, dispatcher, "same", domain.FindElementQuery{Text: "OK"}, PriorityNormal, time.Second, time.Millisecond)
	assertNoStart(t, runner.started)

	runner.release <- nil
	assertNoError(t, <-running)
	waitStarted(t, runner.started, "dump-ui:same")

	runner.release <- nil
	assertNoError(t, <-pendingWait)
}

func TestObservationDispatcherHighPriorityWaitOvertakesPendingNormal(t *testing.T) {
	runner := newBlockingObservationRunner()
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	running := runObservationCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, runner.started, "screenshot:same")

	pendingNormal := runObservationDumpUI(ctx, dispatcher, "same", PriorityNormal)
	pendingHigh := runObservationWaitForElement(ctx, dispatcher, "same", domain.FindElementQuery{Text: "OK"}, PriorityHigh, time.Second, time.Millisecond)
	waitObservationQueueLengths(t, dispatcher, "same", 1, 1)

	runner.release <- nil
	assertNoError(t, <-running)
	waitStarted(t, runner.started, "dump-ui:same")

	runner.release <- nil
	assertNoError(t, <-pendingHigh)
	assertNotDone(t, pendingNormal)

	waitStarted(t, runner.started, "dump-ui:same")
	runner.release <- nil
	assertNoError(t, <-pendingNormal)
}

func TestObservationDispatcherSerializesDetectStateWithOtherTasks(t *testing.T) {
	runner := newBlockingObservationRunner()
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	running := runObservationCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, runner.started, "screenshot:same")

	pendingDetect := runObservationDetectState(ctx, dispatcher, "same", domain.DetectStateOptions{Mode: domain.DetectionModeUI}, PriorityHigh)
	assertNoStart(t, runner.started)

	runner.release <- nil
	assertNoError(t, <-running)
	waitStarted(t, runner.started, "dump-ui:same")

	runner.release <- nil
	assertNoError(t, <-pendingDetect)
}

func TestObservationDispatcherDetectStateCanUseVLMWhenUIDumpFails(t *testing.T) {
	runner := &failingDumpObservationRunner{}
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)
	dispatcher.SetScreenAnalyzer(fakeScreenAnalyzer{
		result: domain.VLMAnalysis{
			State:       "feed",
			Confidence:  0.91,
			BackendUsed: "fake_vlm",
			Description: "Fresh screenshot feed",
			Elements:    []string{"Mdesings", "Underground bunker diy"},
		},
	})

	result, err := dispatcher.DetectState(context.Background(), "same", domain.DetectStateOptions{
		Mode:          domain.DetectionModeVLM,
		Platform:      "tiktok",
		UseScreenshot: true,
	}, PriorityNormal)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != domain.ScreenStateMainFeed || result.BackendUsed != "fake_vlm" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.VLMError != "" {
		t.Fatalf("unexpected vlm error: %q", result.VLMError)
	}
}

func TestObservationDispatcherCurrentScreenAndUIUpdateCache(t *testing.T) {
	runner := &sequenceObservationRunner{
		dumps: []domain.UIDump{
			{Serial: "stub", XMLDump: "<hierarchy/>", ElementCount: 1, TakenAt: time.Now().UTC()},
		},
	}
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	shot, err := dispatcher.CurrentScreen(context.Background(), "stub", PriorityNormal)
	if err != nil {
		t.Fatal(err)
	}
	if shot.Serial != "stub" || shot.ObjectKey == "" {
		t.Fatalf("unexpected screenshot: %+v", shot)
	}
	dump, err := dispatcher.CurrentUI(context.Background(), "stub", PriorityNormal)
	if err != nil {
		t.Fatal(err)
	}
	if dump.Serial != "stub" || dump.ElementCount != 1 {
		t.Fatalf("unexpected dump: %+v", dump)
	}

	cleared, err := dispatcher.ClearCache(context.Background(), "stub", PriorityHigh)
	if err != nil {
		t.Fatal(err)
	}
	if !cleared.Cleared || !cleared.ScreenCleared || !cleared.UICleared || cleared.Serial != "stub" {
		t.Fatalf("unexpected clear result: %+v", cleared)
	}

	cleared, err = dispatcher.ClearCache(context.Background(), "stub", PriorityHigh)
	if err != nil {
		t.Fatal(err)
	}
	if !cleared.Cleared || cleared.ScreenCleared || cleared.UICleared {
		t.Fatalf("second clear should be idempotent: %+v", cleared)
	}
}

func TestObservationDispatcherDoesNotUpdateCacheAfterFailedScreen(t *testing.T) {
	runner := &sequenceObservationRunner{captureErr: domain.ErrScreenshotFailed}
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	_, err := dispatcher.CurrentScreen(context.Background(), "stub", PriorityNormal)
	if !errors.Is(err, domain.ErrScreenshotFailed) {
		t.Fatalf("expected ErrScreenshotFailed, got %v", err)
	}
	cleared, err := dispatcher.ClearCache(context.Background(), "stub", PriorityHigh)
	if err != nil {
		t.Fatal(err)
	}
	if cleared.ScreenCleared || cleared.UICleared {
		t.Fatalf("failed capture must not populate cache: %+v", cleared)
	}
}

func TestObservationDispatcherClearCacheSerializesWithRunningTask(t *testing.T) {
	runner := newBlockingObservationRunner()
	dispatcher := NewObservationDispatcher(runner, runner, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	running := runObservationCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, runner.started, "screenshot:same")

	clearDone := runObservationClearCache(ctx, dispatcher, "same", PriorityHigh)
	assertNotDone(t, clearDone)

	runner.release <- nil
	assertNoError(t, <-running)
	assertNoError(t, <-clearDone)
}

func runObservationCapture(ctx context.Context, dispatcher *ObservationDispatcher, serial string, priority ScreenshotPriority) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := dispatcher.Capture(ctx, serial, priority)
		done <- err
	}()
	return done
}

func runObservationDumpUI(ctx context.Context, dispatcher *ObservationDispatcher, serial string, priority ScreenshotPriority) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := dispatcher.DumpUI(ctx, serial, priority)
		done <- err
	}()
	return done
}

func runObservationFindElement(ctx context.Context, dispatcher *ObservationDispatcher, serial string, query domain.FindElementQuery, priority ScreenshotPriority) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := dispatcher.FindElement(ctx, serial, query, priority)
		done <- err
	}()
	return done
}

func runObservationWaitForElement(ctx context.Context, dispatcher *ObservationDispatcher, serial string, query domain.FindElementQuery, priority ScreenshotPriority, waitTimeout, checkInterval time.Duration) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := dispatcher.WaitForElement(ctx, serial, query, priority, waitTimeout, checkInterval)
		done <- err
	}()
	return done
}

func runObservationDetectState(ctx context.Context, dispatcher *ObservationDispatcher, serial string, options domain.DetectStateOptions, priority ScreenshotPriority) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := dispatcher.DetectState(ctx, serial, options, priority)
		done <- err
	}()
	return done
}

func runObservationClearCache(ctx context.Context, dispatcher *ObservationDispatcher, serial string, priority ScreenshotPriority) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := dispatcher.ClearCache(ctx, serial, priority)
		done <- err
	}()
	return done
}

type sequenceObservationRunner struct {
	dumps      []domain.UIDump
	calls      int
	captureErr error
}

type failingDumpObservationRunner struct{}

func (f *failingDumpObservationRunner) CaptureAndStore(context.Context, string) (domain.Screenshot, error) {
	return domain.Screenshot{Bytes: []byte("png"), TakenAt: time.Now().UTC(), StorageURL: "noop://shot.png", ObjectKey: "shot.png"}, nil
}

func (f *failingDumpObservationRunner) Capture(context.Context, string) (domain.Screenshot, error) {
	return domain.Screenshot{Bytes: []byte("png"), TakenAt: time.Now().UTC()}, nil
}

func (f *failingDumpObservationRunner) DumpUIDocument(context.Context, string) (domain.UIDump, error) {
	return domain.UIDump{}, domain.ErrUIDumpFailed
}

type fakeScreenAnalyzer struct {
	result domain.VLMAnalysis
	err    error
}

func (f fakeScreenAnalyzer) Analyze(context.Context, string, string, []byte) (domain.VLMAnalysis, error) {
	return f.result, f.err
}

func (s *sequenceObservationRunner) CaptureAndStore(context.Context, string) (domain.Screenshot, error) {
	if s.captureErr != nil {
		return domain.Screenshot{}, s.captureErr
	}
	return domain.Screenshot{
		Serial:     "stub",
		ObjectKey:  "stub/shot.png",
		StorageURL: "af-screenshots/stub/shot.png",
		SizeBytes:  67,
		Width:      1,
		Height:     1,
		TakenAt:    time.Now().UTC(),
	}, nil
}

func (s *sequenceObservationRunner) Capture(context.Context, string) (domain.Screenshot, error) {
	return domain.Screenshot{Bytes: []byte("png"), TakenAt: time.Now().UTC()}, nil
}

func (s *sequenceObservationRunner) DumpUIDocument(context.Context, string) (domain.UIDump, error) {
	if s.calls >= len(s.dumps) {
		s.calls++
		return s.dumps[len(s.dumps)-1], nil
	}
	dump := s.dumps[s.calls]
	s.calls++
	return dump, nil
}

func waitObservationQueueLengths(t *testing.T, dispatcher *ObservationDispatcher, serial string, normal, high int) {
	t.Helper()
	deadline := time.After(250 * time.Millisecond)
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()

	for {
		dispatcher.mu.Lock()
		worker := dispatcher.workers[serial]
		dispatcher.mu.Unlock()
		if worker != nil {
			worker.mu.Lock()
			gotNormal := len(worker.normalQueue)
			gotHigh := len(worker.highQueue)
			worker.mu.Unlock()
			if gotNormal == normal && gotHigh == high {
				return
			}
		}

		select {
		case <-deadline:
			t.Fatalf("expected queue lengths normal=%d high=%d for serial %q", normal, high, serial)
		case <-tick.C:
		}
	}
}
