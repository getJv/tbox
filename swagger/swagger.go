package swagger

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
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

	// Serve Swagger UI
	mux.Get(cfg.UIPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		data := struct {
			JSONURL string
		}{
			JSONURL: cfg.JSONPath,
		}
		if err := tmpl.Execute(w, data); err != nil {
			l.Error().Err(err).Msg("failed to execute swagger template")
		}
	})

	// Serve assets
	assetPrefix := path.Join(cfg.UIPath, "assets")
	if !strings.HasSuffix(assetPrefix, "/") {
		assetPrefix += "/"
	}
	subFS, err := fs.Sub(assets, "assets")
	if err != nil {
		return fmt.Errorf("failed to create sub filesystem for assets: %w", err)
	}
	fileServer := http.FileServer(http.FS(subFS))
	mux.Get(assetPrefix+"*", http.StripPrefix(assetPrefix, fileServer).ServeHTTP)

	l.Info().
		Str("url", fmt.Sprintf("http://%s%s", cfg.Host, cfg.UIPath)).
		Msg("Swagger UI available at")

	return nil
}
