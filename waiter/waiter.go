package waiter

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// WaitFunc defines a long-running task that should stop when ctx is canceled.
type WaitFunc func(ctx context.Context) error

// CleanupFunc defines a cleanup callback executed when Wait returns.
type CleanupFunc func()

// Waiter coordinates concurrent wait functions and graceful shutdown.
type Waiter interface {
	// Add registers one or more functions to be run and waited for.
	Add(fns ...WaitFunc)
	// Cleanup registers one or more cleanup callbacks.
	//
	// Cleanup callbacks are deferred and executed when Wait returns.
	Cleanup(fns ...CleanupFunc)
	// Wait runs all registered wait functions and blocks until they finish
	// or the internal context is canceled.
	Wait() error
	// Context returns the internal context shared by the waiter.
	Context() context.Context
	// CancelFunc returns the internal cancel function.
	CancelFunc() context.CancelFunc
}

type waiter struct {
	ctx          context.Context
	waitFuncs    []WaitFunc
	cleanupFuncs []CleanupFunc
	cancel       context.CancelFunc
}

type waiterCfg struct {
	parentCtx    context.Context
	catchSignals bool
}

// New creates a waiter instance with optional configuration.
//
// By default, it uses context.Background() as parent context and does not
// subscribe to OS signals.
func New(options ...WaiterOption) Waiter {
	cfg := &waiterCfg{
		parentCtx:    context.Background(),
		catchSignals: false,
	}

	for _, option := range options {
		option(cfg)
	}

	w := &waiter{
		waitFuncs:    []WaitFunc{},
		cleanupFuncs: []CleanupFunc{},
	}
	w.ctx, w.cancel = context.WithCancel(cfg.parentCtx)
	if cfg.catchSignals {
		w.ctx, w.cancel = signal.NotifyContext(w.ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	}

	return w
}

// Add registers wait functions to be executed by Wait.
func (w *waiter) Add(fns ...WaitFunc) {
	w.waitFuncs = append(w.waitFuncs, fns...)
}

// Cleanup registers cleanup callbacks to be deferred in Wait.
func (w *waiter) Cleanup(fns ...CleanupFunc) {
	w.cleanupFuncs = append(w.cleanupFuncs, fns...)
}

// Wait executes all registered wait functions and waits for completion.
//
// It cancels the internal context when it is done and returns the first
// non-nil error from wait functions, if any.
func (w *waiter) Wait() (err error) {
	g, ctx := errgroup.WithContext(w.ctx)

	g.Go(func() error {
		<-ctx.Done()
		w.cancel()
		return nil
	})

	for _, fn := range w.waitFuncs {
		waitFunc := fn
		g.Go(func() error { return waitFunc(ctx) })
	}

	for _, fn := range w.cleanupFuncs {
		cleanupFunc := fn
		defer cleanupFunc()
	}

	return g.Wait()
}

// Context returns the waiter's internal context.
func (w *waiter) Context() context.Context {
	return w.ctx
}

// CancelFunc returns the waiter's cancel function.
func (w *waiter) CancelFunc() context.CancelFunc {
	return w.cancel
}
