package config

// OpenAIConfig holds the configuration for OpenAI API.
type OpenAIConfig struct {
	APIKey string `envconfig:"OPENAI_API_KEY"`
}
