## di

`di` is a small dependency injection package for registering and resolving services by key.

It supports two lifecycles:

- `Singleton`: one shared instance for the whole application.
- `Scoped`: one instance per request/scope context.

### Installation

```bash
go get github.com/getjv/tbox/di
```

### Main types

- `Container`: main interface used to register and resolve dependencies.
- `DepFactoryFunc`: factory function that receives a `Container` and returns the built dependency.
- `Scope`: lifecycle type (`Singleton` or `Scoped`).

### Methods

#### `New(l zerolog.Logger) Container`

Creates a new root DI container.

#### `AddSingleton(key string, fn DepFactoryFunc)`

Registers a dependency that is built once and reused.

#### `AddScoped(key string, fn DepFactoryFunc)`

Registers a dependency that is built once for each scope.

#### `Scoped(ctx context.Context) context.Context`

Creates a scoped container and stores it in a new context.

#### `Get(key string) any`

Resolves a dependency by key from the container instance.

#### `Get(ctx context.Context, key string) any`

Helper function that resolves a dependency from the scoped container stored in `ctx`.

### Behavior notes

- The package detects cyclic dependency chains and panics with the tracked chain.
- `Get` panics when the key is not registered.
- Singleton dependencies are always resolved from the root container.
- Scoped dependencies are cached only inside their own scope.