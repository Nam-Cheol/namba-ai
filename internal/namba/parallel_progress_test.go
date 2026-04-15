package namba

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type parallelProgressRecord map[string]any

func TestRunParallelProgressEventsCaptureLifecycleOrderingAndExecutionPhases(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-028"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err != nil {
		t.Fatalf("runParallel failed: %v", err)
	}

	events := mustReadParallelProgressEvents(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-028-parallel.events.jsonl"))
	if len(events) == 0 {
		t.Fatal("expected parallel progress events to be written")
	}

	wantPhases := []string{"queued", "running", "validating", "merge_pending", "merging", "done"}
	phasePos := make(map[string]int, len(wantPhases))
	runID := ""
	for i, event := range events {
		if got := eventString(event, "spec_id"); got != "SPEC-028" {
			t.Fatalf("event %d spec_id = %q, want SPEC-028: %+v", i, got, event)
		}
		if got := eventString(event, "source"); got != "lifecycle" {
			t.Fatalf("event %d source = %q, want lifecycle: %+v", i, got, event)
		}
		if got := eventString(event, "scope"); got != "run" && got != "worker" {
			t.Fatalf("event %d scope = %q, want run or worker: %+v", i, got, event)
		}
		if got := eventString(event, "phase"); got == "" {
			t.Fatalf("event %d missing phase: %+v", i, event)
		}
		if got := eventString(event, "status"); got == "" {
			t.Fatalf("event %d missing status: %+v", i, event)
		}
		if got := eventString(event, "timestamp"); got == "" {
			t.Fatalf("event %d missing timestamp: %+v", i, event)
		}
		if got := eventString(event, "summary", "detail"); got == "" {
			t.Fatalf("event %d missing summary/detail: %+v", i, event)
		}

		if seq, ok := eventInt(event, "sequence"); !ok {
			t.Fatalf("event %d missing sequence: %+v", i, event)
		} else if seq != i+1 {
			t.Fatalf("event %d sequence = %d, want %d: %+v", i, seq, i+1, event)
		}

		if got := eventString(event, "run_id"); got == "" {
			t.Fatalf("event %d missing run id: %+v", i, event)
		} else if runID == "" {
			runID = got
		} else if got != runID {
			t.Fatalf("event %d run id = %q, want %q: %+v", i, got, runID, event)
		}

		phase := eventString(event, "phase")
		if _, ok := phasePos[phase]; !ok {
			phasePos[phase] = i
		}
		if eventString(event, "scope") == "worker" {
			if got := eventString(event, "worker_name", "worker_id", "worker"); got == "" {
				t.Fatalf("event %d missing worker identity: %+v", i, event)
			}
		}
	}

	for _, phase := range wantPhases {
		pos, ok := phasePos[phase]
		if !ok {
			t.Fatalf("expected phase %q in progress events, got %+v", phase, collectProgressPhases(events))
		}
		phasePos[phase] = pos
	}
	for i := 1; i < len(wantPhases); i++ {
		if phasePos[wantPhases[i-1]] >= phasePos[wantPhases[i]] {
			t.Fatalf("expected phase %q before %q, got %+v", wantPhases[i-1], wantPhases[i], collectProgressPhases(events))
		}
	}

	if hasProgressPhase(events, "failed") {
		t.Fatalf("did not expect failed phase in successful parallel run, got %+v", collectProgressPhases(events))
	}
}

func TestRunParallelProgressEventsCaptureMergeBlockedAndPreservedFailureDetails(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	failedDir := filepath.Join(h.tmp, worktreesDir, "spec-028-p2")
	h.execFailures[failedDir] = errors.New("runner boom")

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-028"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err == nil {
		t.Fatal("expected execution failure")
	}
	if !strings.Contains(err.Error(), "blocked merge") {
		t.Fatalf("expected blocked merge error, got %v", err)
	}

	events := mustReadParallelProgressEvents(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-028-parallel.events.jsonl"))
	if len(events) == 0 {
		t.Fatal("expected progress events to be preserved on failure")
	}

	if !hasProgressPhase(events, "failed") {
		t.Fatalf("expected failed phase in progress stream, got %+v", collectProgressPhases(events))
	}
	if hasProgressPhase(events, "merging") {
		t.Fatalf("did not expect merge phase after worker failure, got %+v", collectProgressPhases(events))
	}

	preservedEvent := firstProgressEvent(events, func(event parallelProgressRecord) bool {
		return eventString(event, "scope") == "worker" && eventString(event, "phase") == "done" && eventString(event, "status") == "preserved"
	})
	if preservedEvent == nil {
		t.Fatalf("expected preserved worker progress event, got %+v", collectProgressStatuses(events))
	}
	if got := eventString(*preservedEvent, "summary", "detail"); !strings.Contains(strings.ToLower(got), "preserv") && !strings.Contains(strings.ToLower(got), "block") {
		t.Fatalf("expected preserved or blocked detail on preserved worker event, got %q from %+v", got, *preservedEvent)
	}
}

func TestParallelProgressEventParserAcceptsFutureSources(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "spec-028-parallel.events.jsonl")
	line := `{"spec_id":"SPEC-028","run_id":"run-1","sequence":1,"timestamp":"2026-04-15T00:00:00Z","source":"codex-background-agent","scope":"worker","worker_id":"worker-1","phase":"running","status":"streaming","summary":"worker started"}`
	if err := os.WriteFile(path, []byte(line+"\n"), 0o644); err != nil {
		t.Fatalf("write synthetic progress event: %v", err)
	}

	events := mustReadParallelProgressEvents(t, path)
	if len(events) != 1 {
		t.Fatalf("expected one synthetic event, got %d", len(events))
	}
	if got := eventString(events[0], "source"); got != "codex-background-agent" {
		t.Fatalf("expected future source to round-trip, got %q", got)
	}
	if got := eventString(events[0], "worker_name", "worker_id", "worker"); got != "worker-1" {
		t.Fatalf("expected worker identity to round-trip, got %q", got)
	}
}

func TestRunParallelProgressEventsFailFastWhenEventLogPathCannotBeInitialized(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	blockedPath := filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-028-parallel.events.jsonl")
	if err := os.MkdirAll(blockedPath, 0o755); err != nil {
		t.Fatalf("prepare blocked event path: %v", err)
	}

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-028"},
		[]string{"one", "two"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 2},
		false,
	)
	if err == nil {
		t.Fatal("expected progress sink initialization failure")
	}
	if countCommandsContaining(h.commands, "git worktree add -b") != 0 {
		t.Fatalf("expected no worker setup after progress sink initialization failure, got %v", h.commands)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-028-parallel.json"))
	if !report.MergeBlocked || !report.ProgressLogFailed || report.FinishedAt == "" {
		t.Fatalf("expected terminal report with progress log failure, got %+v", report)
	}
}

func TestRunParallelProgressEventsPreserveWorkersWhenAppendFailsDuringExecution(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.app.newParallelProgressSink = func(cfg parallelProgressSinkConfig) (parallelProgressSink, error) {
		return &stubParallelProgressSink{
			path: cfg.Path,
			failMatch: func(input parallelProgressEventInput, publishCount int) bool {
				return input.Scope == parallelProgressScopeWorker && input.Phase == "merge_pending"
			},
			failPublishErr: errors.New("append denied"),
		}, nil
	}

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-028"},
		[]string{"one"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 1},
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "append denied") {
		t.Fatalf("expected append failure to surface, got %v", err)
	}
	if countCommandsContaining(h.commands, "git merge --no-ff") != 0 {
		t.Fatalf("expected no merges after append failure, got %v", h.commands)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-028-parallel.json"))
	if !report.MergeBlocked || !report.ProgressLogFailed || !strings.Contains(report.ProgressLogError, "append denied") {
		t.Fatalf("expected merge-blocked report with progress log failure, got %+v", report)
	}
	if len(report.Workers) != 1 {
		t.Fatalf("expected one worker in report, got %+v", report.Workers)
	}
	worker := report.Workers[0]
	if !worker.Preserved || !worker.ExecutionPassed || !worker.ValidationPassed {
		t.Fatalf("expected preserved worker with successful execution and validation, got %+v", worker)
	}
	if worker.ExecutionError != "" || worker.ValidationError != "" {
		t.Fatalf("expected progress failure to avoid polluting worker failure fields, got %+v", worker)
	}
}

func TestRunParallelProgressEventsRecordMergeFailureBeforePreservedState(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.mergeErr["namba/spec-028-p2"] = errors.New("merge conflict")

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-028"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "merge namba/spec-028-p2") {
		t.Fatalf("expected merge failure to surface, got %v", err)
	}

	events := mustReadParallelProgressEvents(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-028-parallel.events.jsonl"))
	mergeFailedIdx := indexProgressEvent(events, func(event parallelProgressRecord) bool {
		return eventString(event, "scope") == parallelProgressScopeWorker &&
			eventString(event, "worker_name", "worker_id", "worker") == "spec-028-p2" &&
			eventString(event, "phase") == "failed" &&
			eventString(event, "status") == "merge_failed"
	})
	preservedIdx := indexProgressEvent(events, func(event parallelProgressRecord) bool {
		return eventString(event, "scope") == parallelProgressScopeWorker &&
			eventString(event, "worker_name", "worker_id", "worker") == "spec-028-p2" &&
			eventString(event, "phase") == "done" &&
			eventString(event, "status") == "preserved"
	})
	if mergeFailedIdx == -1 || preservedIdx == -1 {
		t.Fatalf("expected merge_failed and preserved events for failed worker, got %+v", events)
	}
	if mergeFailedIdx >= preservedIdx {
		t.Fatalf("expected merge_failed before preserved for failed worker, got %+v", events)
	}
}

func TestRunParallelProgressEventsSurfaceCloseFailureAfterSuccessfulRun(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.app.newParallelProgressSink = func(cfg parallelProgressSinkConfig) (parallelProgressSink, error) {
		return &stubParallelProgressSink{
			path:     cfg.Path,
			closeErr: errors.New("close denied"),
		}, nil
	}

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-028"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "close denied") {
		t.Fatalf("expected close failure to surface, got %v", err)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-028-parallel.json"))
	if report.MergeBlocked || !report.ProgressLogFailed || !strings.Contains(report.ProgressLogError, "close denied") {
		t.Fatalf("expected successful merge report with close failure surfaced, got %+v", report)
	}
	for _, worker := range report.Workers {
		if !worker.MergeSucceeded {
			t.Fatalf("expected merges to complete before close failure, got %+v", worker)
		}
	}
}

func mustReadParallelProgressEvents(t *testing.T, path string) []parallelProgressRecord {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open progress events: %v", err)
	}
	defer file.Close()

	events := make([]parallelProgressRecord, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		event := make(parallelProgressRecord)
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("unmarshal progress event %q: %v", line, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan progress events: %v", err)
	}
	return events
}

func eventString(event parallelProgressRecord, keys ...string) string {
	for _, key := range keys {
		raw, ok := event[key]
		if !ok || raw == nil {
			continue
		}
		switch value := raw.(type) {
		case string:
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		case fmt.Stringer:
			if trimmed := strings.TrimSpace(value.String()); trimmed != "" {
				return trimmed
			}
		case json.Number:
			if trimmed := strings.TrimSpace(value.String()); trimmed != "" {
				return trimmed
			}
		default:
			if text := strings.TrimSpace(fmt.Sprint(value)); text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func eventInt(event parallelProgressRecord, keys ...string) (int, bool) {
	for _, key := range keys {
		raw, ok := event[key]
		if !ok || raw == nil {
			continue
		}
		switch value := raw.(type) {
		case float64:
			return int(value), true
		case float32:
			return int(value), true
		case int:
			return value, true
		case int64:
			return int(value), true
		case json.Number:
			n, err := strconv.Atoi(value.String())
			if err == nil {
				return n, true
			}
		case string:
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err == nil {
				return n, true
			}
		}
	}
	return 0, false
}

func collectProgressPhases(events []parallelProgressRecord) []string {
	phases := make([]string, 0, len(events))
	for _, event := range events {
		phases = append(phases, eventString(event, "phase"))
	}
	return phases
}

func collectProgressStatuses(events []parallelProgressRecord) []string {
	statuses := make([]string, 0, len(events))
	for _, event := range events {
		statuses = append(statuses, eventString(event, "status"))
	}
	return statuses
}

func hasProgressPhase(events []parallelProgressRecord, phase string) bool {
	for _, event := range events {
		if eventString(event, "phase") == phase {
			return true
		}
	}
	return false
}

func firstProgressEvent(events []parallelProgressRecord, match func(parallelProgressRecord) bool) *parallelProgressRecord {
	for i := range events {
		if match(events[i]) {
			return &events[i]
		}
	}
	return nil
}

func indexProgressEvent(events []parallelProgressRecord, match func(parallelProgressRecord) bool) int {
	for i := range events {
		if match(events[i]) {
			return i
		}
	}
	return -1
}

type stubParallelProgressSink struct {
	mu             sync.Mutex
	path           string
	failAtPublish  int
	failMatch      func(parallelProgressEventInput, int) bool
	publishCount   int
	failPublishErr error
	closeErr       error
	writeErr       error
}

func (s *stubParallelProgressSink) Publish(input parallelProgressEventInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.writeErr != nil {
		return s.writeErr
	}
	s.publishCount++
	shouldFail := s.failAtPublish > 0 && s.publishCount >= s.failAtPublish
	if !shouldFail && s.failMatch != nil {
		shouldFail = s.failMatch(input, s.publishCount)
	}
	if shouldFail {
		s.writeErr = wrapStubParallelProgressError("append", s.path, s.failPublishErr)
		return s.writeErr
	}
	return nil
}

func (s *stubParallelProgressSink) Close() error {
	if s.closeErr == nil {
		return nil
	}
	return wrapStubParallelProgressError("close", s.path, s.closeErr)
}

func (s *stubParallelProgressSink) Path() string {
	return s.path
}

func wrapStubParallelProgressError(stage, path string, err error) error {
	if err == nil {
		err = errors.New("stub progress error")
	}
	var progressErr *parallelProgressError
	if errors.As(err, &progressErr) {
		return err
	}
	return &parallelProgressError{Stage: stage, Path: path, Err: err}
}
