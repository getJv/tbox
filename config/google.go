package config

// GoogleConfig holds the configuration for Google OAuth2.
type GoogleConfig struct {
	clientID     string `envconfig:"GOOGLE_CLIENT_ID"`
	clientSecret string `envconfig:"GOOGLE_CLIENT_SECRET"`
	redirectURL  string `envconfig:"AUTH_REDIRECT_URL"`
}

// ClientID returns the Google client ID.
func (g GoogleConfig) ClientID() string {
	return g.clientID
}

// ClientSecret returns the Google client secret.
func (g GoogleConfig) ClientSecret() string {
	return g.clientSecret
}

// RedirectURL returns the Google redirect URL.
func (g GoogleConfig) RedirectURL() string {
	return g.redirectURL
}
