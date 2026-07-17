package config

// GoogleConfig holds the configuration for Google OAuth2.
type GoogleConfig struct {
	ClientID     string `envconfig:"GOOGLE_CLIENT_ID"`
	ClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET"`
	RedirectURL  string `envconfig:"AUTH_REDIRECT_URL"`
}
