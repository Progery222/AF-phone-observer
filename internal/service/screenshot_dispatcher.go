package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mobilefarm/af/phone-observer/internal/domain"
	"github.com/mobilefarm/af/phone-observer/internal/port"
)

const (
	DefaultScreenshotQueueSize     = 32
	DefaultScreenshotHighQueueSize = 8
)

type ScreenshotPriority string

const (
	PriorityNormal ScreenshotPriority = "normal"
	PriorityHigh   ScreenshotPriority = "high"
)

var (
	ErrInvalidPriority = errors.New("неизвестный приоритет задачи")
	ErrQueueFull       = errors.New("очередь скриншотов заполнена")
)

type ScreenshotCapturer interface {
	Capture(ctx context.Context, serial string) (domain.Screenshot, error)
	CaptureAndStore(ctx context.Context, serial string) (domain.Screenshot, error)
}

type UIDumpCapturer interface {
	DumpUIDocument(ctx context.Context, serial string) (domain.UIDump, error)
}

type ObservationDispatcher struct {
	screenshots     ScreenshotCapturer
	uiDumps         UIDumpCapturer
	analyzer        port.ScreenAnalyzer
	normalQueueSize int
	highQueueSize   int

	mu      sync.Mutex
	workers map[string]*observationWorker

	cacheMu sync.Mutex
	cache   map[string]observationCache
}

type ScreenshotDispatcher = ObservationDispatcher

func NewObservationDispatcher(screenshots ScreenshotCapturer, uiDumps UIDumpCapturer, normalQueueSize, highQueueSize int) *ObservationDispatcher {
	if normalQueueSize <= 0 {
		normalQueueSize = DefaultScreenshotQueueSize
	}
	if highQueueSize <= 0 {
		highQueueSize = DefaultScreenshotHighQueueSize
	}
	return &ObservationDispatcher{
		screenshots:     screenshots,
		uiDumps:         uiDumps,
		normalQueueSize: normalQueueSize,
		highQueueSize:   highQueueSize,
		workers:         make(map[string]*observationWorker),
		cache:           make(map[string]observationCache),
	}
}

func NewScreenshotDispatcher(capturer ScreenshotCapturer, normalQueueSize, highQueueSize int) *ScreenshotDispatcher {
	return NewObservationDispatcher(capturer, nil, normalQueueSize, highQueueSize)
}

func (d *ObservationDispatcher) SetScreenAnalyzer(analyzer port.ScreenAnalyzer) {
	d.analyzer = analyzer
}

func ParseScreenshotPriority(raw string) (ScreenshotPriority, error) {
	switch ScreenshotPriority(raw) {
	case "", PriorityNormal:
		return PriorityNormal, nil
	case PriorityHigh:
		return PriorityHigh, nil
	default:
		return "", ErrInvalidPriority
	}
}

func (d *ObservationDispatcher) Capture(ctx context.Context, serial string, priority ScreenshotPriority) (domain.Screenshot, error) {
	if d.screenshots == nil {
		return domain.Screenshot{}, domain.ErrScreenshotFailed
	}
	result, err := d.dispatch(ctx, serial, priority, func(ctx context.Context, serial string) observationResult {
		shot, err := d.screenshots.CaptureAndStore(ctx, serial)
		if err == nil {
			d.storeScreenshotCache(serial, shot)
		}
		return observationResult{screenshot: shot, err: err}
	})
	if err != nil {
		return domain.Screenshot{}, err
	}
	return result.screenshot, result.err
}

func (d *ObservationDispatcher) CurrentScreen(ctx context.Context, serial string, priority ScreenshotPriority) (domain.Screenshot, error) {
	return d.Capture(ctx, serial, priority)
}

func (d *ObservationDispatcher) DumpUI(ctx context.Context, serial string, priority ScreenshotPriority) (domain.UIDump, error) {
	if d.uiDumps == nil {
		return domain.UIDump{}, domain.ErrUIDumpFailed
	}
	result, err := d.dispatch(ctx, serial, priority, func(ctx context.Context, serial string) observationResult {
		dump, err := d.uiDumps.DumpUIDocument(ctx, serial)
		if err == nil {
			d.storeUIDumpCache(serial, dump)
		}
		return observationResult{uiDump: dump, err: err}
	})
	if err != nil {
		return domain.UIDump{}, err
	}
	return result.uiDump, result.err
}

func (d *ObservationDispatcher) CurrentUI(ctx context.Context, serial string, priority ScreenshotPriority) (domain.UIDump, error) {
	return d.DumpUI(ctx, serial, priority)
}

func (d *ObservationDispatcher) FindElement(ctx context.Context, serial string, query domain.FindElementQuery, priority ScreenshotPriority) (domain.FindElementResult, error) {
	if d.uiDumps == nil {
		return domain.FindElementResult{}, domain.ErrUIDumpFailed
	}
	if err := domain.ValidateFindElementQuery(query); err != nil {
		return domain.FindElementResult{}, err
	}
	result, err := d.dispatch(ctx, serial, priority, func(ctx context.Context, serial string) observationResult {
		dump, err := d.uiDumps.DumpUIDocument(ctx, serial)
		if err != nil {
			return observationResult{err: err}
		}
		d.storeUIDumpCache(serial, dump)
		found, err := domain.FindElement(dump.Elements, query)
		return observationResult{findElement: found, err: err}
	})
	if err != nil {
		return domain.FindElementResult{}, err
	}
	return result.findElement, result.err
}

func (d *ObservationDispatcher) ClearCache(ctx context.Context, serial string, priority ScreenshotPriority) (domain.CacheClearResult, error) {
	result, err := d.dispatch(ctx, serial, priority, func(_ context.Context, serial string) observationResult {
		return observationResult{clearCache: d.clearCache(serial)}
	})
	if err != nil {
		return domain.CacheClearResult{}, err
	}
	return result.clearCache, result.err
}

func (d *ObservationDispatcher) WaitForElement(ctx context.Context, serial string, query domain.FindElementQuery, priority ScreenshotPriority, waitTimeout, checkInterval time.Duration) (domain.WaitForElementResult, error) {
	if d.uiDumps == nil {
		return domain.WaitForElementResult{}, domain.ErrUIDumpFailed
	}
	if err := domain.ValidateFindElementQuery(query); err != nil {
		return domain.WaitForElementResult{}, err
	}
	result, err := d.dispatch(ctx, serial, priority, func(ctx context.Context, serial string) observationResult {
		waitResult, err := d.waitForElement(ctx, serial, query, waitTimeout, checkInterval)
		return observationResult{waitElement: waitResult, err: err}
	})
	if err != nil {
		return domain.WaitForElementResult{}, err
	}
	return result.waitElement, result.err
}

func (d *ObservationDispatcher) DetectState(ctx context.Context, serial string, options domain.DetectStateOptions, priority ScreenshotPriority) (domain.ScreenDetection, error) {
	if d.uiDumps == nil {
		return domain.ScreenDetection{}, domain.ErrUIDumpFailed
	}
	mode, err := domain.NormalizeDetectionMode(options.Mode)
	if err != nil {
		return domain.ScreenDetection{}, err
	}
	if mode == domain.DetectionModeVLM && !options.UseScreenshot {
		return domain.ScreenDetection{}, domain.ErrInvalidDetectMode
	}
	options.Mode = mode
	options.Platform = domain.NormalizeDetectionPlatform(options.Platform)

	result, err := d.dispatch(ctx, serial, priority, func(ctx context.Context, serial string) observationResult {
		detection, err := d.detectState(ctx, serial, options)
		return observationResult{detection: detection, err: err}
	})
	if err != nil {
		return domain.ScreenDetection{}, err
	}
	return result.detection, result.err
}

func (d *ObservationDispatcher) detectState(ctx context.Context, serial string, options domain.DetectStateOptions) (domain.ScreenDetection, error) {
	dump, err := d.uiDumps.DumpUIDocument(ctx, serial)
	if err != nil {
		if options.Mode == domain.DetectionModeUI || d.analyzer == nil || !options.UseScreenshot {
			return domain.ScreenDetection{}, err
		}
		return d.detectStateFromScreenshotOnly(ctx, serial, options)
	}
	d.storeUIDumpCache(serial, dump)
	detection := domain.DetectScreenFromUIDump(dump)
	if options.Mode == domain.DetectionModeUI {
		return detection, nil
	}
	if options.Mode == domain.DetectionModeVLM && d.analyzer == nil {
		return domain.ScreenDetection{}, domain.ErrVLMUnavailable
	}

	needsVLM := options.UseScreenshot && d.analyzer != nil
	needsScreenshot := needsVLM || options.StoreScreenshot
	var shot domain.Screenshot
	if needsScreenshot {
		if d.screenshots == nil {
			return domain.ScreenDetection{}, domain.ErrScreenshotFailed
		}
		if options.StoreScreenshot {
			shot, err = d.screenshots.CaptureAndStore(ctx, serial)
		} else {
			shot, err = d.screenshots.Capture(ctx, serial)
		}
		if err != nil {
			return domain.ScreenDetection{}, err
		}
		d.storeScreenshotCache(serial, shot)
		applyScreenshotToDetection(&detection, shot)
	}

	if options.Mode == domain.DetectionModeVLM && !needsVLM {
		return domain.ScreenDetection{}, domain.ErrVLMUnavailable
	}
	if !needsVLM {
		return detection, nil
	}

	vlm, err := d.analyzer.Analyze(ctx, serial, options.Platform, shot.Bytes)
	if err != nil {
		if options.Mode == domain.DetectionModeVLM {
			return domain.ScreenDetection{}, err
		}
		detection.VLMError = err.Error()
		return detection, nil
	}
	detection = domain.MergeScreenDetections(detection, vlm)
	applyScreenshotToDetection(&detection, shot)
	return detection, nil
}

func (d *ObservationDispatcher) detectStateFromScreenshotOnly(ctx context.Context, serial string, options domain.DetectStateOptions) (domain.ScreenDetection, error) {
	if d.screenshots == nil {
		return domain.ScreenDetection{}, domain.ErrScreenshotFailed
	}
	var (
		shot domain.Screenshot
		err  error
	)
	if options.StoreScreenshot {
		shot, err = d.screenshots.CaptureAndStore(ctx, serial)
	} else {
		shot, err = d.screenshots.Capture(ctx, serial)
	}
	if err != nil {
		return domain.ScreenDetection{}, err
	}
	d.storeScreenshotCache(serial, shot)
	vlm, err := d.analyzer.Analyze(ctx, serial, options.Platform, shot.Bytes)
	if err != nil {
		return domain.ScreenDetection{}, err
	}
	detection := domain.MergeScreenDetections(domain.ScreenDetection{
		Serial:         serial,
		State:          domain.ScreenStateUnknown,
		Confidence:     0,
		Source:         "vlm",
		BackendUsed:    "vlm",
		Description:    "UI dump недоступен, состояние определено по screenshot",
		Elements:       []string{},
		MatchedSignals: []string{},
		TakenAt:        shot.TakenAt,
	}, vlm)
	applyScreenshotToDetection(&detection, shot)
	return detection, nil
}

func applyScreenshotToDetection(detection *domain.ScreenDetection, shot domain.Screenshot) {
	if shot.StorageURL != "" {
		detection.ScreenshotURL = shot.StorageURL
	}
	if shot.ObjectKey != "" {
		detection.MinioKey = shot.ObjectKey
	}
	if detection.TakenAt.IsZero() && !shot.TakenAt.IsZero() {
		detection.TakenAt = shot.TakenAt
	}
}

func (d *ObservationDispatcher) waitForElement(ctx context.Context, serial string, query domain.FindElementQuery, waitTimeout, checkInterval time.Duration) (domain.WaitForElementResult, error) {
	startedAt := time.Now()
	deadline := startedAt.Add(waitTimeout)
	checkCount := 0
	for {
		if err := ctx.Err(); err != nil {
			return domain.WaitForElementResult{}, err
		}
		if !time.Now().Before(deadline) {
			return waitTimeoutResult(startedAt, waitTimeout, checkCount), domain.ErrElementWaitTimeout
		}

		dump, err := d.uiDumps.DumpUIDocument(ctx, serial)
		if err != nil {
			return domain.WaitForElementResult{}, err
		}
		d.storeUIDumpCache(serial, dump)
		checkCount++

		found, err := domain.FindElement(dump.Elements, query)
		if err == nil {
			return domain.WaitForElementResult{
				Found:      true,
				Element:    found.Element,
				FoundBy:    found.FoundBy,
				WaitTimeMS: time.Since(startedAt).Milliseconds(),
				CheckCount: checkCount,
			}, nil
		}
		if !errors.Is(err, domain.ErrElementNotFound) {
			return domain.WaitForElementResult{}, err
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return waitTimeoutResult(startedAt, waitTimeout, checkCount), domain.ErrElementWaitTimeout
		}
		sleepFor := checkInterval
		if sleepFor > remaining {
			sleepFor = remaining
		}
		timer := time.NewTimer(sleepFor)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return domain.WaitForElementResult{}, ctx.Err()
		}
	}
}

func waitTimeoutResult(startedAt time.Time, waitTimeout time.Duration, checkCount int) domain.WaitForElementResult {
	return domain.WaitForElementResult{
		Found:      false,
		FoundBy:    "",
		TimeoutSec: int(waitTimeout.Seconds()),
		WaitTimeMS: time.Since(startedAt).Milliseconds(),
		CheckCount: checkCount,
	}
}

func (d *ObservationDispatcher) dispatch(ctx context.Context, serial string, priority ScreenshotPriority, run observationRunFunc) (observationResult, error) {
	if serial == "" {
		return observationResult{}, domain.ErrInvalidSerial
	}
	priority, err := ParseScreenshotPriority(string(priority))
	if err != nil {
		return observationResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return observationResult{}, err
	}

	job := observationJob{
		ctx:    ctx,
		serial: serial,
		run:    run,
		result: make(chan observationResult, 1),
	}
	worker := d.workerFor(serial)
	if err := worker.enqueue(job, priority); err != nil {
		return observationResult{}, err
	}

	select {
	case result := <-job.result:
		return result, nil
	case <-ctx.Done():
		return observationResult{}, ctx.Err()
	}
}

func (d *ObservationDispatcher) workerFor(serial string) *observationWorker {
	d.mu.Lock()
	defer d.mu.Unlock()

	worker := d.workers[serial]
	if worker != nil {
		return worker
	}
	worker = newObservationWorker(serial, d.normalQueueSize, d.highQueueSize)
	d.workers[serial] = worker
	return worker
}

func (d *ObservationDispatcher) storeScreenshotCache(serial string, shot domain.Screenshot) {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()

	entry := d.cache[serial]
	entry.screenshot = shot
	entry.hasScreenshot = true
	d.cache[serial] = entry
}

func (d *ObservationDispatcher) storeUIDumpCache(serial string, dump domain.UIDump) {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()

	entry := d.cache[serial]
	entry.uiDump = dump
	entry.hasUIDump = true
	d.cache[serial] = entry
}

func (d *ObservationDispatcher) clearCache(serial string) domain.CacheClearResult {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()

	entry := d.cache[serial]
	delete(d.cache, serial)
	return domain.CacheClearResult{
		Serial:        serial,
		Cleared:       true,
		ScreenCleared: entry.hasScreenshot,
		UICleared:     entry.hasUIDump,
		ClearedAt:     time.Now().UTC(),
	}
}

type observationCache struct {
	screenshot    domain.Screenshot
	uiDump        domain.UIDump
	hasScreenshot bool
	hasUIDump     bool
}

type observationWorker struct {
	serial      string
	normalLimit int
	highLimit   int

	mu          sync.Mutex
	cond        *sync.Cond
	normalQueue []observationJob
	highQueue   []observationJob
}

type observationRunFunc func(ctx context.Context, serial string) observationResult

type observationJob struct {
	ctx    context.Context
	serial string
	run    observationRunFunc
	result chan observationResult
}

type observationResult struct {
	screenshot  domain.Screenshot
	uiDump      domain.UIDump
	findElement domain.FindElementResult
	waitElement domain.WaitForElementResult
	detection   domain.ScreenDetection
	clearCache  domain.CacheClearResult
	err         error
}

func newObservationWorker(serial string, normalLimit, highLimit int) *observationWorker {
	worker := &observationWorker{
		serial:      serial,
		normalLimit: normalLimit,
		highLimit:   highLimit,
	}
	worker.cond = sync.NewCond(&worker.mu)
	go worker.run()
	return worker
}

func (w *observationWorker) enqueue(job observationJob, priority ScreenshotPriority) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	switch priority {
	case PriorityHigh:
		if len(w.highQueue) >= w.highLimit {
			return ErrQueueFull
		}
		w.highQueue = append(w.highQueue, job)
	case PriorityNormal:
		if len(w.normalQueue) >= w.normalLimit {
			return ErrQueueFull
		}
		w.normalQueue = append(w.normalQueue, job)
	default:
		return ErrInvalidPriority
	}
	w.cond.Signal()
	return nil
}

func (w *observationWorker) run() {
	for {
		job := w.next()
		if err := job.ctx.Err(); err != nil {
			job.result <- observationResult{err: err}
			continue
		}
		job.result <- job.run(job.ctx, job.serial)
	}
}

func (w *observationWorker) next() observationJob {
	w.mu.Lock()
	defer w.mu.Unlock()

	for len(w.highQueue) == 0 && len(w.normalQueue) == 0 {
		w.cond.Wait()
	}
	if len(w.highQueue) > 0 {
		job := w.highQueue[0]
		w.highQueue[0] = observationJob{}
		w.highQueue = w.highQueue[1:]
		return job
	}
	job := w.normalQueue[0]
	w.normalQueue[0] = observationJob{}
	w.normalQueue = w.normalQueue[1:]
	return job
}
