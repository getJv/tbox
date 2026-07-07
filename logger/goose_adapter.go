package logger

import "github.com/rs/zerolog"

// GooseAdapter adapta o zerolog para a interface esperada pelo goose.
type GooseAdapter struct {
	Log zerolog.Logger
}

func (a *GooseAdapter) Fatalf(format string, v ...interface{}) { a.Log.Fatal().Msgf(format, v...) }
func (a *GooseAdapter) Printf(format string, v ...interface{}) { a.Log.Info().Msgf(format, v...) }
