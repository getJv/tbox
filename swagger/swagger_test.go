package swagger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

type mockRouter struct {
	routes map[string]http.HandlerFunc
}

func (m *mockRouter) Get(path string, handler http.HandlerFunc) {
	if m.routes == nil {
		m.routes = make(map[string]http.HandlerFunc)
	}
	m.routes[path] = handler
}

func TestRegister_Assets(t *testing.T) {
	mux := &mockRouter{}
	l := zerolog.New(nil)
	cfg := Config{
		UIPath:   "/api/docs/",
		JSONPath: "/api/docs/swagger.json",
		FilePath: "test.json",
		Host:     "localhost",
	}

	err := Register(mux, l, cfg)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	assetPath := "/api/docs/assets/swagger-ui.css"
	handler, ok := mux.routes["/api/docs/assets/*"]
	if !ok {
		t.Fatalf("Asset handler not registered at /api/docs/assets/*")
	}

	req := httptest.NewRequest("GET", assetPath, nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/css; charset=utf-8" && contentType != "text/css" {
		t.Errorf("Expected Content-Type text/css, got %q", contentType)
	}
}

func TestRegister_JSAssets(t *testing.T) {
	mux := &mockRouter{}
	l := zerolog.New(nil)
	cfg := Config{
		UIPath:   "/api/docs/",
		JSONPath: "/api/docs/swagger.json",
		FilePath: "test.json",
		Host:     "localhost",
	}

	err := Register(mux, l, cfg)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	assets := []string{
		"swagger-ui-bundle.js",
		"swagger-ui-standalone-preset.js",
	}

	handler, ok := mux.routes["/api/docs/assets/*"]
	if !ok {
		t.Fatalf("Asset handler not registered at /api/docs/assets/*")
	}

	for _, asset := range assets {
		t.Run(asset, func(t *testing.T) {
			assetPath := "/api/docs/assets/" + asset
			req := httptest.NewRequest("GET", assetPath, nil)
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType != "text/javascript; charset=utf-8" && contentType != "application/javascript" && contentType != "text/javascript" {
				t.Errorf("Expected Content-Type javascript, got %q", contentType)
			}
		})
	}
}
