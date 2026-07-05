package log

import (
	std "log"
)

type SilentLogger struct{}

// Fatal logs the message and calls os.Exit(1).
func (*SilentLogger) Fatal(v ...any) { std.Fatal(v...) }

// Fatalf logs the formatted message and calls os.Exit(1).
func (*SilentLogger) Fatalf(format string, v ...any) { std.Fatalf(format, v...) }

// Print ignores the message.
func (*SilentLogger) Print(...any) {
	// ignores the message on purpose
}

// Println ignores the message.
func (*SilentLogger) Println(...any) {
	// ignores the message on purpose
}

// Printf ignores the message.
func (*SilentLogger) Printf(string, ...any) {
	// ignores the message on purpose
}
