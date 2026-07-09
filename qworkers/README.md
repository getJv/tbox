# qworkers

`qworkers` is a simple, SQL-based task queue package for Go, designed to handle background jobs using SQLite or other SQL databases as the backing store.

## Features

- **Persistence**: Jobs are stored in a database, ensuring they are not lost if the process restarts.
- **Retries**: Configurable maximum retries with a default delay between attempts.
- **Atomicity**: Uses SQL transactions to ensure jobs are picked up by only one worker at a time.
- **Decoupled Handlers**: Register handlers for different queues easily.

## Installation and Migrations

This package uses `goose` for database migrations. The migrations are embedded within the package.

To ensure the necessary tables exist, you can use the `SchemaProvider` which implements a `MigrationProvider` interface:

```go
// SchemaProvider ensures the required tables for the worker exist in the database.
type SchemaProvider struct{}

func (p *SchemaProvider) EnsureSchema(db *sql.DB) error {
    return RunMigration(db, "up")
}
```

In your system initialization, you should run these migrations:

```go
for _, provider := range s.migrationProviders {
    if err := provider.EnsureSchema(s.db); err != nil {
        return fmt.Errorf("failed to run provider migration: %w", err)
    }
}
```

Alternatively, you can run migrations manually using the `RunMigration` function provided by the package.

## Usage

### 1. Initialize the Repository and Worker

```go
func (s *System) initQueues() {
    // Initialize the repository (e.g., SQLite)
    queueRepo := queueinfra.NewSQLiteQueueRepository(s.DB())
    
    // Create the worker
    s.workers = qworkers.NewWorker(queueRepo, s.logger)
}
```

### 2. Register Handlers

Register a handler for a specific queue name. The handler will be called with the job's payload.

```go
err := s.workers.RegisterHandler("email_queue", func(ctx context.Context, payload string) error {
    // Process the job
    fmt.Printf("Processing email: %s\n", payload)
    return nil
})
```

### 3. Start the Worker

The worker polls the registered queues for new jobs.

```go
func (s *System) WaitForQueueWorkers(ctx context.Context) error {
    return s.workers.Start(ctx)
}
```

## Database Schema

The default implementation uses a `queue_jobs` table to manage tasks. See the `migrations` directory for specific SQL definitions.

