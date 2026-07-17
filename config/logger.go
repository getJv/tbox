package config

import "github.com/getjv/tbox/logger"

// LoggerConfig holds the configuration for the logger.
type LoggerConfig struct {
	Loglevel     string `envconfig:"LOG_LEVEL" default:"INFO"`
	LogDirectory string `envconfig:"LOG_DIRECTORY" default:"./logs"`
	LogFilename  string `envconfig:"LOG_FILENAME" default:"app.log"`
}

// Directory returns the log directory.
func (l LoggerConfig) Directory() string {
	return l.LogDirectory
}

// Level returns the logger level.
func (l LoggerConfig) Level() logger.LoggerLevel {
	return logger.LoggerLevel(l.Loglevel)
}

// Filename returns the log filename.
func (l LoggerConfig) Filename() string {
	return l.LogFilename
}
