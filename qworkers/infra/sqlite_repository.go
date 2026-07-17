package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/getjv/tbox/qworkers"
)

const defaultRetryDelay = 1 * time.Minute

// sqliteRepository persists queue jobs in an SQLite database.
type sqliteRepository struct {
	db *sql.DB
}

// NewSQLiteQueueRepository creates a new instance of a SQLite-backed queue repository.
func NewSQLiteQueueRepository(db *sql.DB) qworkers.Repository {
	return &sqliteRepository{db: db}
}

// Enqueue inserts a new pending job into the database for the specified queue.
func (r *sqliteRepository) Enqueue(ctx context.Context, queueName string, payload []byte) error {
	const query = `
		INSERT INTO queue_jobs (queue_name, payload, status, retries, max_retries, run_at, created_at, updated_at)
		VALUES (?, ?, ?, 0, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := r.db.ExecContext(ctx, query, queueName, payload, qworkers.JobStatusPending, qworkers.DefaultMaxRetries)
	if err != nil {
		return fmt.Errorf("enqueue queue job: %w", err)
	}

	return nil
}

// PickNextJob atomically identifies and moves the next available pending job to the processing status.
// It includes a retry mechanism for handling potential concurrent access conflicts.
func (r *sqliteRepository) PickNextJob(ctx context.Context, queueName string) (*qworkers.Job, error) {
	for range 3 {
		job, err := r.pickNextJobOnce(ctx, queueName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}

			if errors.Is(err, errConcurrentPick) {
				continue
			}
			return nil, err
		}

		return job, nil
	}

	return nil, nil
}

func (r *sqliteRepository) pickNextJobOnce(ctx context.Context, queueName string) (*qworkers.Job, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx to pick queue job: %w", err)
	}

	job, err := r.pickNextInTx(ctx, tx, queueName)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, fmt.Errorf("rollback pick queue job tx: %w", rollbackErr)
		}

		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit pick queue job tx: %w", err)
	}

	return job, nil
}

// MarkAsCompleted updates a job's status to completed in the database.
func (r *sqliteRepository) MarkAsCompleted(ctx context.Context, id int64) error {
	const query = `UPDATE queue_jobs SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, qworkers.JobStatusCompleted, id)
	if err != nil {
		return fmt.Errorf("mark queue job as completed: %w", err)
	}

	return nil
}

// MarkAsFailed increments the retry count and either schedules a retry or marks the job as permanently failed.
func (r *sqliteRepository) MarkAsFailed(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx to fail queue job: %w", err)
	}

	var retries int
	var maxRetries int
	if err := tx.QueryRowContext(ctx, `SELECT retries, max_retries FROM queue_jobs WHERE id = ?`, id).Scan(&retries, &maxRetries); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("load queue job retries: %w", err)
	}

	nextRetries := retries + 1
	nextStatus := qworkers.JobStatusPending
	nextRunAt := time.Now().UTC().Add(defaultRetryDelay)
	if nextRetries >= maxRetries {
		nextStatus = qworkers.JobStatusFailed
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE queue_jobs SET status = ?, retries = ?, run_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		nextStatus,
		nextRetries,
		nextRunAt,
		id,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("update failed queue job: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit fail queue job tx: %w", err)
	}

	return nil
}

var errConcurrentPick = errors.New("concurrent queue pick")

// pickNextInTx selects and locks a single pending job within the provided transaction.
func (r *sqliteRepository) pickNextInTx(ctx context.Context, tx *sql.Tx, queueName string) (*qworkers.Job, error) {
	var id int64
	err := tx.QueryRowContext(
		ctx,
		`SELECT id FROM queue_jobs WHERE queue_name = ? AND status = ? AND run_at <= CURRENT_TIMESTAMP ORDER BY run_at ASC, id ASC LIMIT 1`,
		queueName,
		qworkers.JobStatusPending,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(
		ctx,
		`UPDATE queue_jobs SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = ?`,
		qworkers.JobStatusProcessing,
		id,
		qworkers.JobStatusPending,
	)
	if err != nil {
		return nil, fmt.Errorf("set queue job processing: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read queue job update rows: %w", err)
	}
	if affected == 0 {
		return nil, errConcurrentPick
	}

	job := &qworkers.Job{}
	if err := tx.QueryRowContext(
		ctx,
		`SELECT id, queue_name, payload, status, retries, max_retries, run_at FROM queue_jobs WHERE id = ?`,
		id,
	).Scan(&job.ID, &job.QueueName, &job.Payload, &job.Status, &job.Retries, &job.MaxRetries, &job.RunAt); err != nil {
		return nil, fmt.Errorf("load picked queue job: %w", err)
	}

	return job, nil
}
