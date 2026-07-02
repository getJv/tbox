package di

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// modeled after this excellent DI lib: https://github.com/sarulabs/di

// Scope defines the lifecycle of a dependency in the container.
type Scope int

const (
	// Singleton creates one shared instance in the root container.
	Singleton Scope = iota + 1
	// Scoped creates one instance per scoped container.
	Scoped
)

type contextKey int

const containerKey contextKey = 1

// DepFactoryFunc builds a dependency using the current container.
type DepFactoryFunc func(c Container) (any, error)

type tempValue = chan struct{}

// Container defines the dependency injection API.
type Container interface {
	// AddSingleton registers a singleton dependency factory.
	AddSingleton(key string, fn DepFactoryFunc)
	// AddScoped registers a scoped dependency factory.
	AddScoped(key string, fn DepFactoryFunc)
	// Scoped returns a new context that contains a scoped container.
	Scoped(ctx context.Context) context.Context
	// Get resolves a dependency by key.
	Get(key string) any
}

// depInfo stores metadata used to build a dependency.
type depInfo struct {
	key     string
	scope   Scope
	factory DepFactoryFunc
}

var _ Container = (*container)(nil)

// container is the default implementation of Container.
type container struct {
	parent  *container
	deps    map[string]depInfo
	vals    map[string]any
	tracked tracked
	logger  zerolog.Logger
	mu      sync.Mutex
}

// New creates a dependency container with the provided logger.
func New(l zerolog.Logger) Container {

	diContainer := &container{
		logger: l,
		deps:   make(map[string]depInfo),
		vals:   make(map[string]any),
	}

	diContainer.logger.Info().Msg("Initializing dependency injection container...")

	return diContainer
}

// AddSingleton registers a dependency that is built once and reused.
func (c *container) AddSingleton(key string, fn DepFactoryFunc) {
	c.deps[key] = depInfo{
		key:     key,
		scope:   Singleton,
		factory: fn,
	}
	c.logger.Info().Msgf("Registered dependency: %s, scope: singleton", key)
}

// AddScoped registers a dependency that is built once per scope.
func (c *container) AddScoped(key string, fn DepFactoryFunc) {
	c.deps[key] = depInfo{
		key:     key,
		scope:   Scoped,
		factory: fn,
	}
}

// Scoped creates a child scope and stores it in a returned context.
func (c *container) Scoped(ctx context.Context) context.Context {
	return context.WithValue(ctx, containerKey, c.scoped())
}

// Get resolves a dependency by key.
//
// It panics when the key is not registered or when a cyclic dependency is found.
func (c *container) Get(key string) any {
	info, exists := c.deps[key]
	if !exists {
		panic(fmt.Sprintf("there is no dependency registered with `%s`", key))
	}

	// catch cases of: building Foo needs Bar and building Bar needs Foo :boom:
	if _, exists := c.tracked[info.key]; exists {
		panic(fmt.Sprintf("cyclic dependencies encountered while building `%s`, tracked: %s", info.key, c.tracked))
	}

	if info.scope == Singleton {
		return c.getFromParent(info)
	}

	return c.get(info)
}

// getFromParent resolves singleton dependencies from the root container.
func (c *container) getFromParent(info depInfo) any {
	if c.parent != nil {
		return c.parent.getFromParent(info)
	}

	return c.get(info)
}

// get returns a cached value or builds it once in a concurrent-safe way.
func (c *container) get(info depInfo) any {
	c.mu.Lock()

	v, exists := c.vals[info.key]
	if !exists {
		tv := make(tempValue)
		c.vals[info.key] = tv
		c.mu.Unlock()
		return c.build(info, tv)
	}

	c.mu.Unlock()
	tv, isTemp := v.(tempValue)
	if !isTemp {
		return v
	}

	<-tv

	return c.get(info)
}

// build executes a dependency factory and stores the built value.
func (c *container) build(info depInfo, tv tempValue) any {
	v, err := info.factory(c.builder(info))

	c.mu.Lock()

	if err != nil {
		delete(c.vals, info.key)
		c.mu.Unlock()
		close(tv)
		panic(fmt.Sprintf("error building dependency `%s`: %s", info.key, err))
	}

	c.vals[info.key] = v
	c.mu.Unlock()
	close(tv)

	return v
}

// scoped creates a child container that shares registrations with the parent.
func (c *container) scoped() *container {
	return &container{
		parent: c,
		deps:   c.deps,
		vals:   make(map[string]any),
	}
}

// builder creates a helper container used while building a dependency.
func (c *container) builder(info depInfo) *container {
	return &container{
		parent:  c.parent,
		deps:    c.deps,
		vals:    c.vals,
		tracked: c.tracked.add(info),
	}
}
