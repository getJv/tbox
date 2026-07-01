## waiter

`waiter` is a small utility package to coordinate long-running goroutines and shutdown flow in Go applications.

It helps you:

- Share a single context across workers.
- Wait for all registered workers to finish.
- Cancel everything on OS signals (graceful shutdown).
- Run cleanup callbacks when waiting ends.

### Installation

```bash
go get github.com/getjv/tbox/waiter
```

### Core concepts

- `WaitFunc`: a function that receives a context and returns an error.
- `CleanupFunc`: a callback executed when `Wait()` returns.
- `Waiter`: coordinator that stores tasks, exposes context/cancel, and blocks in `Wait()`.

### Best practices

- Always use the context provided by `w.Context()` (or the one passed into each `WaitFunc`).
- Make each worker stop on `<-ctx.Done()`.
- Return meaningful errors from workers to surface failures quickly.
- Use `CatchSignals()` in applications to handle `SIGINT`/`SIGTERM` gracefully.
- Register resource cleanup (close DB, flush logs, stop servers) with `Cleanup()`.

### Example

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/getjv/tbox/waiter"
)

func runHTTPServer(ctx context.Context, srv *http.Server) error {
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})

	return g.Wait()
}

func main() {
	w := waiter.New(
		waiter.ParentContext(context.Background()),
		waiter.CatchSignals(),
	)

	srv := &http.Server{Addr: ":8080"}

	w.Add(
		func(ctx context.Context) error {
			return runHTTPServer(ctx, srv)
		},
		func(ctx context.Context) error {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					fmt.Println("health-check: alive")
				}
			}
		},
	)

	w.Cleanup(
		func() {
			fmt.Println("cleanup: releasing resources")
		},
	)

	if err := w.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown with error: %v\n", err)
		os.Exit(1)
	}
}
```

### Behavior notes

- `Wait()` starts all registered `WaitFunc` items concurrently.
- If one worker returns an error, the shared context is canceled and the error is returned.
- Cleanup callbacks are executed with `defer` inside `Wait()` (last registered, first executed).