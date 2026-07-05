package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

const (
	maxSizeMB  = 200
	maxBackups = 3
	maxAgeDays = 7
)

// LoggerLevel represents the severity level of the log message.
type LoggerLevel string

const (
	// TRACE is the most verbose level.
	TRACE LoggerLevel = "TRACE"
	// DEBUG provides detailed information for debugging.
	DEBUG LoggerLevel = "DEBUG"
	// INFO provides general informational messages.
	INFO LoggerLevel = "INFO"
	// WARN indicates potential issues.
	WARN LoggerLevel = "WARN"
	// ERROR indicates a failure in a specific operation.
	ERROR LoggerLevel = "ERROR"
	// PANIC indicates a critical failure that causes the application to exit.
	PANIC LoggerLevel = "PANIC"
)

// LoggerConfig defines the interface for configuring the logger.
type LoggerConfig interface {
	// Directory returns the directory where log files will be stored.
	Directory() string
	// Filename returns the name of the log file.
	Filename() string
	// Level returns the logging severity level.
	Level() LoggerLevel
}

// New creates and initializes a new zerolog.Logger instance.
// In development environment, it clears the log file on startup.
// It writes logs to both the console and a rotating log file.
func New(env string, cfg LoggerConfig) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(cfg.Directory(), 0755); err != nil {
		// If we can't create the directory, log to stderr and fallback
		log.Printf("Failed to create log directory: %v", err)
		return createConsoleLogger(env, cfg)
	}

	logPath := filepath.Join(cfg.Directory(), cfg.Filename())

	// Clear logs on startup in development
	if env == "development" {
		if f, err := os.Create(logPath); err == nil {
			if err := f.Close(); err != nil {
				log.Printf("Failed to close log file: %v", err)
			}
		}
	}

	// Create a multi-writer that writes to both console and file
	fileWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSizeMB, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays, // days
		Compress:   true,
	}

	// Create console writer for terminal output
	consoleWriter := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = "03:04:05.000PM"
	})

	// Multi-writer to write to both console and file
	multiWriter := io.MultiWriter(consoleWriter, fileWriter)

	return zerolog.New(multiWriter).
		Level(logLevelToZero(cfg.Level())).
		With().
		Timestamp().
		Logger()
}

func createConsoleLogger(env string, cfg LoggerConfig) zerolog.Logger {
	switch env {
	case "production":
		return zerolog.New(os.Stdout).
			Level(logLevelToZero(cfg.Level())).
			With().
			Timestamp().
			Logger()
	default:
		return zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = "03:04:05.000PM"
		})).
			Level(logLevelToZero(cfg.Level())).
			With().
			Timestamp().
			Logger()
	}
}

func logLevelToZero(level LoggerLevel) zerolog.Level {
	switch level {
	case PANIC:
		return zerolog.PanicLevel
	case ERROR:
		return zerolog.ErrorLevel
	case WARN:
		return zerolog.WarnLevel
	case INFO:
		return zerolog.InfoLevel
	case DEBUG:
		return zerolog.DebugLevel
	case TRACE:
		return zerolog.TraceLevel
	default:
		return zerolog.InfoLevel
	}
}
