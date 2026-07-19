package config

// OpenAIConfig holds the configuration for OpenAI API.
type OpenAIConfig struct {
	APIKey  string `envconfig:"OPENAI_API_KEY"`
	BaseURL string `envconfig:"OPENAI_BASE_URL" default:"https://api.openai.com/v1/chat/completions"`
	Model   string `envconfig:"OPENAI_MODEL" default:"gpt-4o-mini"`
}

func (c OpenAIConfig) GetOpenAIAPIKey() string  { return c.APIKey }
func (c OpenAIConfig) GetOpenAIBaseURL() string { return c.BaseURL }
func (c OpenAIConfig) GetOpenAIModel() string   { return c.Model }
