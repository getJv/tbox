# Logger

The `logger` package provides a structured logging utility for the tbox project, featuring console output and file rotation.

## Features

- **Multi-writer**: Logs to both console and file simultaneously.
- **Log Rotation**: Automatically rotates log files based on size, age, and number of backups.
- **Environment Aware**: Clears logs on startup in development mode.
- **Structured**: Powered by `zerolog` for high-performance structured logging.

## Usage

### Configuration

Implement the `LoggerConfig` interface to provide the necessary settings:

```go
type Config struct {
    // ... fields
}

func (c Config) Directory() string { return "logs" }
func (c Config) Filename() string  { return "tbox.log" }
func (c Config) Level() logger.LoggerLevel { return logger.DEBUG }
```

### Initialization

Use the `New` function to create a logger instance:

```go
cfg := Config{}
l := logger.New("development", cfg)

l.Info().Msg("hello world")
```

## Silent Logger

The `log` subpackage provides a `SilentLogger` struct that can be used to suppress log output while maintaining compatibility with standard logging interfaces.
