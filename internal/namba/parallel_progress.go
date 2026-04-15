package namba

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	parallelProgressSourceLifecycle = "lifecycle"
	parallelProgressScopeRun        = "run"
	parallelProgressScopeWorker     = "worker"
)

type parallelProgressEventInput struct {
	Source     string
	Scope      string
	WorkerName string
	Phase      string
	Status     string
	Summary    string
	Detail     string
	Metadata   map[string]any
}

type parallelProgressSink interface {
	Publish(parallelProgressEventInput) error
	Close() error
	Path() string
}

type parallelProgressSinkConfig struct {
	Path   string
	SpecID string
	RunID  string
	Now    func() time.Time
}

type parallelProgressFile interface {
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

type parallelProgressError struct {
	Stage string
	Path  string
	Err   error
}

func (e *parallelProgressError) Error() string {
	return fmt.Sprintf("parallel progress %s %s: %v", e.Stage, e.Path, e.Err)
}

func (e *parallelProgressError) Unwrap() error {
	return e.Err
}

func isParallelProgressFailure(err error) bool {
	var progressErr *parallelProgressError
	return errors.As(err, &progressErr)
}

type parallelProgressEvent struct {
	SpecID     string         `json:"spec_id"`
	RunID      string         `json:"run_id"`
	Sequence   int64          `json:"sequence"`
	Timestamp  string         `json:"timestamp"`
	Source     string         `json:"source"`
	Scope      string         `json:"scope"`
	WorkerName string         `json:"worker_name,omitempty"`
	Phase      string         `json:"phase"`
	Status     string         `json:"status"`
	Summary    string         `json:"summary"`
	Detail     string         `json:"detail,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type jsonlParallelProgressSink struct {
	path     string
	specID   string
	runID    string
	now      func() time.Time
	file     parallelProgressFile
	mu       sync.Mutex
	sequence int64
	writeErr error
	closed   bool
}

func newJSONLParallelProgressSink(cfg parallelProgressSinkConfig) (parallelProgressSink, error) {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return nil, &parallelProgressError{Stage: "initialize", Path: cfg.Path, Err: err}
	}
	file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, &parallelProgressError{Stage: "initialize", Path: cfg.Path, Err: err}
	}
	return &jsonlParallelProgressSink{
		path:   cfg.Path,
		specID: cfg.SpecID,
		runID:  cfg.RunID,
		now:    cfg.Now,
		file:   file,
	}, nil
}

func (s *jsonlParallelProgressSink) Publish(input parallelProgressEventInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.writeErr != nil {
		return s.writeErr
	}
	if s.closed {
		s.writeErr = &parallelProgressError{Stage: "append", Path: s.path, Err: errors.New("sink already closed")}
		return s.writeErr
	}

	s.sequence++
	event := parallelProgressEvent{
		SpecID:     s.specID,
		RunID:      s.runID,
		Sequence:   s.sequence,
		Timestamp:  s.now().Format(time.RFC3339Nano),
		Source:     firstNonBlank(strings.TrimSpace(input.Source), parallelProgressSourceLifecycle),
		Scope:      strings.TrimSpace(input.Scope),
		WorkerName: strings.TrimSpace(input.WorkerName),
		Phase:      strings.TrimSpace(input.Phase),
		Status:     strings.TrimSpace(input.Status),
		Summary:    strings.TrimSpace(input.Summary),
		Detail:     strings.TrimSpace(input.Detail),
	}
	if len(input.Metadata) > 0 {
		event.Metadata = copyParallelProgressMetadata(input.Metadata)
	}

	line, err := json.Marshal(event)
	if err != nil {
		s.writeErr = &parallelProgressError{Stage: "append", Path: s.path, Err: err}
		return s.writeErr
	}
	if _, err := s.file.Write(append(line, '\n')); err != nil {
		s.writeErr = &parallelProgressError{Stage: "append", Path: s.path, Err: err}
		return s.writeErr
	}
	return nil
}

func (s *jsonlParallelProgressSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	var errs []error
	if err := s.file.Sync(); err != nil {
		errs = append(errs, &parallelProgressError{Stage: "close", Path: s.path, Err: err})
	}
	if err := s.file.Close(); err != nil {
		errs = append(errs, &parallelProgressError{Stage: "close", Path: s.path, Err: err})
	}
	return errors.Join(errs...)
}

func (s *jsonlParallelProgressSink) Path() string {
	return s.path
}

func copyParallelProgressMetadata(values map[string]any) map[string]any {
	clone := make(map[string]any, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func newParallelRunID(specID string, now time.Time) string {
	return strings.ToLower(strings.TrimSpace(specID)) + "-parallel-" + now.UTC().Format("20060102t150405.000000000z")
}

func parallelProgressLogPath(root, specID string) string {
	return filepath.Join(root, logsDir, "runs", strings.ToLower(strings.TrimSpace(specID))+"-parallel.events.jsonl")
}

func relativeParallelProgressLogPath(specID string) string {
	return filepath.ToSlash(filepath.Join(logsDir, "runs", strings.ToLower(strings.TrimSpace(specID))+"-parallel.events.jsonl"))
}
