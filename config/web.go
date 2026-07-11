package config

import "fmt"

// WebConfig holds the configuration for the web server.
type WebConfig struct {
	Host              string `envconfig:"HTTP_HOST" default:"0.0.0.0"`
	Port              string `envconfig:"HTTP_PORT" default:"8888"`
	StaticFilesPath   string `envconfig:"HTTP_STATIC_FILES_PATH" default:"./backend/public"`
	StaticPathContext string `envconfig:"HTTP_STATIC_STATIC_PATH_CONTEXT" default:"/"`
}

// Address returns the web server address in host:port format.
func (w WebConfig) Address() string {
	return fmt.Sprintf("%s:%s", w.Host, w.Port)
}
