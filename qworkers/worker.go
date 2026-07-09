package qworkers

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/rs/zerolog"
)

const defaultPollInterval = 1 * time.Second

// Worker pulls jobs from registered queues and dispatches handlers.
type Worker struct {
	repo         Repository
	logger       zerolog.Logger
	handlers     map[string]JobHandler
	pollInterval time.Duration
}

// NewWorker creates a new queue worker instance.
func NewWorker(repo Repository, logger zerolog.Logger) *Worker {
	return &Worker{
		repo:         repo,
		logger:       logger,
		handlers:     make(map[string]JobHandler),
		pollInterval: defaultPollInterval,
	}
}

// RegisterHandler attaches a handler to a specific queue.
// It returns an error if the queue name is empty, the handler is nil, or if a handler
// is already registered for the given queue name.
func (w *Worker) RegisterHandler(queueName string, handler JobHandler) error {
	if queueName == "" {
		return fmt.Errorf("queue name is required")
	}
	if handler == nil {
		return fmt.Errorf("handler for queue %q cannot be nil", queueName)
	}
	if w.HasHandler(queueName) {
		return fmt.Errorf("handler for queue %q already registered", queueName)
	}

	w.handlers[queueName] = handler
	return nil
}

// HasHandler returns true if a handler is registered for the specified queue name.
func (w *Worker) HasHandler(queueName string) bool {
	_, ok := w.handlers[queueName]
	return ok
}

// Start begins polling all registered queues for jobs until the context is canceled.
func (w *Worker) Start(ctx context.Context) error {

	for _, queueName := range slices.Collect(maps.Keys(w.handlers)) {
		if w.HasHandler(queueName) {
			continue
		}

		w.logger.Warn().
			Str("queue_name", queueName).
			Str("worker_state", "missing_handler").
			Msg("Declared queue has no registered handler")
	}

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			w.processTick(ctx)
		}
	}
}

// processTick iterates through all registered queues once and processes any available jobs.
func (w *Worker) processTick(ctx context.Context) {
	for queueName, handler := range w.handlers {
		_ = w.processQueue(ctx, queueName, handler)
	}
}

// processQueue continuously drains available jobs from a specific queue until it is empty.
func (w *Worker) processQueue(ctx context.Context, queueName string, handler JobHandler) error {
	for {
		job, err := w.repo.PickNextJob(ctx, queueName)
		if err != nil {
			return err
		}
		if job == nil {
			return nil
		}

		if err := handler(ctx, job.Payload); err != nil {
			w.logger.Error().
				Err(err).
				Int64("job_id", job.ID).
				Str("queue_name", queueName).
				Msg("Queue job handler failed")
			_ = w.repo.MarkAsFailed(ctx, job.ID)
			continue
		}

		if err := w.repo.MarkAsCompleted(ctx, job.ID); err != nil {
			w.logger.Error().
				Err(err).
				Int64("job_id", job.ID).
				Str("queue_name", queueName).
				Msg("Failed to mark job as completed")
		}
	}
}
