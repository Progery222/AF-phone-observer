package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
)

type blockingScreenshotCapturer struct {
	started chan string
	release chan error
}

func newBlockingScreenshotCapturer() *blockingScreenshotCapturer {
	return &blockingScreenshotCapturer{
		started: make(chan string, 16),
		release: make(chan error, 16),
	}
}

func (b *blockingScreenshotCapturer) CaptureAndStore(ctx context.Context, serial string) (domain.Screenshot, error) {
	b.started <- serial
	select {
	case err := <-b.release:
		if err != nil {
			return domain.Screenshot{}, err
		}
		return domain.Screenshot{
			Serial:     serial,
			ObjectKey:  serial + "/shot.png",
			StorageURL: "af-screenshots/" + serial + "/shot.png",
			TakenAt:    time.Now().UTC(),
		}, nil
	case <-ctx.Done():
		return domain.Screenshot{}, ctx.Err()
	}
}

func (b *blockingScreenshotCapturer) Capture(ctx context.Context, serial string) (domain.Screenshot, error) {
	b.started <- serial
	select {
	case err := <-b.release:
		if err != nil {
			return domain.Screenshot{}, err
		}
		return domain.Screenshot{
			Serial:  serial,
			Bytes:   []byte("png"),
			TakenAt: time.Now().UTC(),
		}, nil
	case <-ctx.Done():
		return domain.Screenshot{}, ctx.Err()
	}
}

func TestScreenshotDispatcherSameSerialSequential(t *testing.T) {
	capturer := newBlockingScreenshotCapturer()
	dispatcher := NewScreenshotDispatcher(capturer, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	firstDone := runDispatcherCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, capturer.started, "same")

	secondDone := runDispatcherCapture(ctx, dispatcher, "same", PriorityNormal)
	assertNoStart(t, capturer.started)

	capturer.release <- nil
	assertNoError(t, <-firstDone)
	waitStarted(t, capturer.started, "same")

	capturer.release <- nil
	assertNoError(t, <-secondDone)
}

func TestScreenshotDispatcherDifferentSerialsParallel(t *testing.T) {
	capturer := newBlockingScreenshotCapturer()
	dispatcher := NewScreenshotDispatcher(capturer, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	firstDone := runDispatcherCapture(ctx, dispatcher, "a", PriorityNormal)
	waitStarted(t, capturer.started, "a")

	secondDone := runDispatcherCapture(ctx, dispatcher, "b", PriorityNormal)
	waitStarted(t, capturer.started, "b")

	capturer.release <- nil
	capturer.release <- nil
	assertNoError(t, <-firstDone)
	assertNoError(t, <-secondDone)
}

func TestScreenshotDispatcherHighPriorityOvertakesPendingNormal(t *testing.T) {
	capturer := newBlockingScreenshotCapturer()
	dispatcher := NewScreenshotDispatcher(capturer, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	runningNormal := runDispatcherCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, capturer.started, "same")

	pendingNormal := runDispatcherCapture(ctx, dispatcher, "same", PriorityNormal)
	pendingHigh := runDispatcherCapture(ctx, dispatcher, "same", PriorityHigh)
	waitQueueLengths(t, dispatcher, "same", 1, 1)

	capturer.release <- nil
	assertNoError(t, <-runningNormal)
	waitStarted(t, capturer.started, "same")

	capturer.release <- nil
	assertNoError(t, <-pendingHigh)
	assertNotDone(t, pendingNormal)

	waitStarted(t, capturer.started, "same")
	capturer.release <- nil
	assertNoError(t, <-pendingNormal)
}

func TestScreenshotDispatcherHighPriorityDoesNotInterruptRunningTask(t *testing.T) {
	capturer := newBlockingScreenshotCapturer()
	dispatcher := NewScreenshotDispatcher(capturer, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	runningNormal := runDispatcherCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, capturer.started, "same")

	pendingHigh := runDispatcherCapture(ctx, dispatcher, "same", PriorityHigh)
	assertNoStart(t, capturer.started)

	capturer.release <- nil
	assertNoError(t, <-runningNormal)
	waitStarted(t, capturer.started, "same")

	capturer.release <- nil
	assertNoError(t, <-pendingHigh)
}

func TestScreenshotDispatcherTakesHighPriorityAfterError(t *testing.T) {
	capturer := newBlockingScreenshotCapturer()
	dispatcher := NewScreenshotDispatcher(capturer, 4, 4)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	runningNormal := runDispatcherCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, capturer.started, "same")

	pendingNormal := runDispatcherCapture(ctx, dispatcher, "same", PriorityNormal)
	pendingHigh := runDispatcherCapture(ctx, dispatcher, "same", PriorityHigh)
	waitQueueLengths(t, dispatcher, "same", 1, 1)

	capturer.release <- domain.ErrScreenshotFailed
	if err := <-runningNormal; !errors.Is(err, domain.ErrScreenshotFailed) {
		t.Fatalf("expected screenshot error, got %v", err)
	}
	waitStarted(t, capturer.started, "same")

	capturer.release <- nil
	assertNoError(t, <-pendingHigh)
	assertNotDone(t, pendingNormal)

	waitStarted(t, capturer.started, "same")
	capturer.release <- nil
	assertNoError(t, <-pendingNormal)
}

func TestScreenshotDispatcherReturnsQueueFull(t *testing.T) {
	capturer := newBlockingScreenshotCapturer()
	dispatcher := NewScreenshotDispatcher(capturer, 1, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	running := runDispatcherCapture(ctx, dispatcher, "same", PriorityNormal)
	waitStarted(t, capturer.started, "same")

	pending := runDispatcherCapture(ctx, dispatcher, "same", PriorityNormal)
	waitQueueLengths(t, dispatcher, "same", 1, 0)

	_, err := dispatcher.Capture(ctx, "same", PriorityNormal)
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}

	capturer.release <- nil
	assertNoError(t, <-running)
	waitStarted(t, capturer.started, "same")
	capturer.release <- nil
	assertNoError(t, <-pending)
}

func runDispatcherCapture(ctx context.Context, dispatcher *ScreenshotDispatcher, serial string, priority ScreenshotPriority) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := dispatcher.Capture(ctx, serial, priority)
		done <- err
	}()
	return done
}

func waitStarted(t *testing.T, started <-chan string, want string) {
	t.Helper()
	select {
	case got := <-started:
		if got != want {
			t.Fatalf("expected serial %q, got %q", want, got)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatalf("expected capture for serial %q to start", want)
	}
}

func waitQueueLengths(t *testing.T, dispatcher *ScreenshotDispatcher, serial string, normal, high int) {
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

func assertNoStart(t *testing.T, started <-chan string) {
	t.Helper()
	select {
	case got := <-started:
		t.Fatalf("capture started unexpectedly for serial %q", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func assertNotDone(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		t.Fatalf("capture finished unexpectedly with %v", err)
	case <-time.After(50 * time.Millisecond):
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
