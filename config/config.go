package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/getjv/tbox/utils"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

// AppConfig represents the application configuration.
type (
	AppConfig struct {
		Environment     string
		ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
		Web             WebConfig
		Logger          LoggerConfig
		OpenAI          OpenAIConfig
		Google          GoogleConfig
		JWTSecret       string `envconfig:"JWT_SECRET" default:"your-secret-here"`
	}
)

// IsDevelopment returns true if the current environment is development.
func (c AppConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// MarshalZerologObject implements the zerolog.LogObjectMarshaler interface to mask sensitive keys in logs.
func (c AppConfig) MarshalZerologObject(e *zerolog.Event) {
	e.Str("Environment", c.Environment).
		Dur("ShutdownTimeout", c.ShutdownTimeout).
		Interface("Web", c.Web).
		Interface("Logger", c.Logger).
		Str("OpenAIKey", utils.MaskString(c.OpenAI.APIKey)).
		Str("GoogleClientID", c.Google.ClientID).
		Str("GoogleClientSecret", utils.MaskString(c.Google.ClientSecret)).
		Str("RedirectURL", c.Google.RedirectURL).
		Str("JWTSecret", utils.MaskString(c.JWTSecret))
}

// InitConfig initializes the application configuration.
func InitConfig() (cfg AppConfig, err error) {
	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = "."
	}

	fullPath := filepath.Join(envPath, ".env")
	err = godotenv.Load(fullPath)
	if err != nil {
		fmt.Printf("Warning: error loading %s: %v\n", fullPath, err)
	}
	fmt.Printf("raw OPENAI_API_KEY: %q\n", os.Getenv("OPENAI_API_KEY"))
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	fmt.Printf("Starting application in environment: %s\n", env)

	_ = godotenv.Overload(filepath.Join(envPath, ".env."+env))

	err = envconfig.Process("", &cfg)
	if err != nil {
		return cfg, err
	}

	cfg.Environment = env
	fmt.Println("Environment configurations loaded")

	return cfg, nil
}
