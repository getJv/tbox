package config

// OpenAIConfig holds the configuration for OpenAI API.
type OpenAIConfig struct {
	apiKey string `envconfig:"OPENAI_API_KEY"`
}

// APIKey returns the OpenAI API key.
func (o OpenAIConfig) APIKey() string {
	return o.apiKey
}
