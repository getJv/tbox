package config

import "github.com/getjv/tbox/logger"

// LoggerConfig holds the configuration for the logger.
type LoggerConfig struct {
	level     string `envconfig:"LOG_LEVEL" default:"INFO"`
	directory string `envconfig:"LOG_DIRECTORY" default:"./logs"`
	filename  string `envconfig:"LOG_FILENAME" default:"app.log"`
}

// Directory returns the log directory.
func (l LoggerConfig) Directory() string {
	return l.directory
}

// Level returns the logger level.
func (l LoggerConfig) Level() logger.LoggerLevel {
	return logger.LoggerLevel(l.level)
}

// Filename returns the log filename.
func (l LoggerConfig) Filename() string {
	return l.filename
}
