package qworkers

import (
	"context"
	"time"
)

const (
	// JobStatusPending indicates the job is waiting to be processed.
	JobStatusPending = "pending"
	// JobStatusProcessing indicates the job is currently being handled by a worker.
	JobStatusProcessing = "processing"
	// JobStatusCompleted indicates the job has successfully finished.
	JobStatusCompleted = "completed"
	// JobStatusFailed indicates the job has failed all retry attempts.
	JobStatusFailed = "failed"
)

const (
	// DefaultMaxRetries is the default number of times a job will be retried before failing.
	DefaultMaxRetries = 3
)

// Job represents one queued task persisted in the database.
type Job struct {
	ID         int64
	QueueName  string
	Payload    string
	Status     string
	Retries    int
	MaxRetries int
	RunAt      time.Time
}

// JobHandler processes one queue payload.
// It should return an error if the job should be retried.
type JobHandler func(ctx context.Context, payload string) error

// Repository defines queue persistence operations.
type Repository interface {
	// Enqueue inserts a new job into the specified queue.
	Enqueue(ctx context.Context, queueName string, payload string) error
	// PickNextJob atomically finds and marks the next available job as processing.
	PickNextJob(ctx context.Context, queueName string) (*Job, error)
	// MarkAsCompleted marks a job as successfully finished.
	MarkAsCompleted(ctx context.Context, id int64) error
	// MarkAsFailed increments retry count and either requeues or fails the job permanently.
	MarkAsFailed(ctx context.Context, id int64) error
}
