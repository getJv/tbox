package swagger

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

//go:embed index.html
var indexTemplate string

//go:embed assets/*
var assets embed.FS

// Router is an interface that defines the required methods for registering Swagger routes.
// Common routers like chi, gin, and echo satisfy this interface.
type Router interface {
	Get(path string, handler http.HandlerFunc)
}

// Config defines the configuration for serving Swagger UI and the OpenAPI specification.
type Config struct {
	// UIPath is the URL path where the Swagger UI will be served (e.g., "/api/docs/").
	UIPath string
	// JSONPath is the URL path where the OpenAPI JSON specification will be served (e.g., "/api/docs/swagger.json").
	JSONPath string
	// FilePath is the local filesystem path to the generated swagger.json file.
	FilePath string
	// Host is the host address used for logging the full Swagger UI URL.
	Host string
	// AssetsPath is the URL path segment where the Swagger UI assets will be served (defaults to "assets").
	AssetsPath string
}

// Register configures the provided router to serve Swagger UI and the OpenAPI JSON specification.
// It uses an embedded HTML template for the Swagger UI and returns an error if the template parsing fails.
func Register(mux Router, l zerolog.Logger, cfg Config) error {
	tmpl, err := template.New("swagger").Parse(indexTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse swagger template: %w", err)
	}

	// Serve swagger.json
	mux.Get(cfg.JSONPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, cfg.FilePath)
	})

	// Normalize UIPath to have a trailing slash for proper relative asset loading
	uiPath := cfg.UIPath
	if !strings.HasSuffix(uiPath, "/") {
		uiPath += "/"
	}

	// Redirect UIPath without trailing slash to UIPath with trailing slash
	uiPathNoSlash := strings.TrimSuffix(uiPath, "/")
	if uiPathNoSlash != "" && uiPathNoSlash != cfg.JSONPath {
		mux.Get(uiPathNoSlash, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, uiPath, http.StatusMovedPermanently)
		})
	}

	// Serve Swagger UI
	mux.Get(uiPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		assetsURL := cfg.AssetsPath
		if assetsURL == "" {
			assetsURL = "assets"
		}
		data := struct {
			JSONURL   string
			AssetsURL string
		}{
			JSONURL:   cfg.JSONPath,
			AssetsURL: assetsURL,
		}
		if err := tmpl.Execute(w, data); err != nil {
			l.Error().Err(err).Msg("failed to execute swagger template")
		}
	})

	// Serve assets
	assetsPath := cfg.AssetsPath
	if assetsPath == "" {
		assetsPath = "assets"
	}
	assetPrefix := uiPath + assetsPath + "/"
	subFS, err := fs.Sub(assets, "assets")
	if err != nil {
		return fmt.Errorf("failed to create sub filesystem for assets: %w", err)
	}
	fileServer := http.FileServer(http.FS(subFS))
	mux.Get(assetPrefix+"*", http.StripPrefix(assetPrefix, fileServer).ServeHTTP)

	l.Info().
		Str("url", fmt.Sprintf("http://%s%s", cfg.Host, uiPath)).
		Msg("Swagger UI available at")

	return nil
}
