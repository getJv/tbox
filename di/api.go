package di

import "context"

// Get resolves a dependency by key from a scoped container stored in the context.
//
// It panics if the context does not contain a DI container.
func Get(ctx context.Context, key string) any {
	ctn, ok := ctx.Value(containerKey).(*container)
	if !ok {
		panic("container does not exist on context")
	}

	return ctn.Get(key)
}
