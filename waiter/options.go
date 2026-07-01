package waiter

import (
	"context"
)

type WaiterOption func(c *waiterCfg)

// ParentContext sets the parent context used to create the waiter context.
//
// If not provided, the waiter uses context.Background() as parent.
func ParentContext(ctx context.Context) WaiterOption {
	return func(c *waiterCfg) {
		c.parentCtx = ctx
	}
}

// CatchSignals enables OS signal handling for graceful shutdown.
//
// When enabled, the waiter context is canceled on interrupt/termination
// signals such as SIGINT, SIGTERM, and SIGQUIT.
func CatchSignals() WaiterOption {
	return func(c *waiterCfg) {
		c.catchSignals = true
	}
}
