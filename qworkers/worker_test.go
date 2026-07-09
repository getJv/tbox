package qworkers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	job            *Job
	pickCalls      int
	completedIDs   []int64
	failedIDs      []int64
	markFailedErr  error
	markSuccessErr error
}

func (f *fakeRepository) Enqueue(ctx context.Context, queueName string, payload string) error {
	return nil
}

func (f *fakeRepository) PickNextJob(ctx context.Context, queueName string) (*Job, error) {
	f.pickCalls++
	if f.pickCalls == 1 {
		return f.job, nil
	}
	return nil, nil
}

func (f *fakeRepository) MarkAsCompleted(ctx context.Context, id int64) error {
	f.completedIDs = append(f.completedIDs, id)
	return f.markSuccessErr
}

func (f *fakeRepository) MarkAsFailed(ctx context.Context, id int64) error {
	f.failedIDs = append(f.failedIDs, id)
	return f.markFailedErr
}

func TestWorker_RegisterAndHasHandler(t *testing.T) {
	repo := &fakeRepository{}
	worker := NewWorker(repo, zerolog.Nop())

	err := worker.RegisterHandler("QueueChapterToVocabulary", func(ctx context.Context, payload string) error {
		return nil
	})
	require.NoError(t, err)
	assert.True(t, worker.HasHandler("QueueChapterToVocabulary"))
	assert.False(t, worker.HasHandler("QueueVocabularyCreation"))
}

func TestWorker_RegisterHandlerRejectsDuplicate(t *testing.T) {
	repo := &fakeRepository{}
	worker := NewWorker(repo, zerolog.Nop())

	err := worker.RegisterHandler("QueueChapterToVocabulary", func(ctx context.Context, payload string) error {
		return nil
	})
	require.NoError(t, err)

	err = worker.RegisterHandler("QueueChapterToVocabulary", func(ctx context.Context, payload string) error {
		return nil
	})
	require.Error(t, err)
}

func TestWorker_StartProcessesSuccessfulJob(t *testing.T) {
	repo := &fakeRepository{
		job: &Job{
			ID:        11,
			QueueName: "QueueChapterToVocabulary",
			Payload:   `{"id":"11"}`,
		},
	}

	worker := NewWorker(repo, zerolog.Nop())
	err := worker.RegisterHandler("QueueChapterToVocabulary", func(ctx context.Context, payload string) error {
		return nil
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()

	err = worker.Start(ctx)
	require.NoError(t, err)
	assert.Equal(t, []int64{11}, repo.completedIDs)
	assert.Empty(t, repo.failedIDs)
}

func TestWorker_StartMarksFailedJob(t *testing.T) {
	repo := &fakeRepository{
		job: &Job{
			ID:        22,
			QueueName: "QueueChapterToVocabulary",
			Payload:   `{"id":"22"}`,
		},
	}

	worker := NewWorker(repo, zerolog.Nop())
	err := worker.RegisterHandler("QueueChapterToVocabulary", func(ctx context.Context, payload string) error {
		return errors.New("handler failure")
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()

	err = worker.Start(ctx)
	require.NoError(t, err)
	assert.Empty(t, repo.completedIDs)
	assert.Equal(t, []int64{22}, repo.failedIDs)
}
