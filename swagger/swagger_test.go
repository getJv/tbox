package swagger

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRegister(t *testing.T) {
	mux := &mockRouter{}
	l := zerolog.Nop()
	cfg := Config{
		UIPath:   "/api/docs/",
		JSONPath: "/api/docs/swagger.json",
		FilePath: "testdata/swagger.json",
		Host:     "localhost",
	}

	err := Register(mux, l, cfg)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	t.Run("Asset serving", func(t *testing.T) {
		assetHandler, ok := mux.routes["/api/docs/assets/*"]
		if !ok {
			t.Fatalf("Asset handler not registered at /api/docs/assets/*")
		}

		assets := []struct {
			name string
			path string
			ct   string
		}{
			{"CSS", "/api/docs/assets/swagger-ui.css", "text/css"},
			{"JS Bundle", "/api/docs/assets/swagger-ui-bundle.js", "javascript"},
			{"JS Preset", "/api/docs/assets/swagger-ui-standalone-preset.js", "javascript"},
		}

		for _, a := range assets {
			t.Run(a.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", a.path, nil)
				w := httptest.NewRecorder()
				assetHandler(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200 for %s, got %d", a.path, resp.StatusCode)
				}
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, a.ct) {
					t.Errorf("Expected Content-Type %s, got %q", a.ct, ct)
				}
			})
		}
	})

	t.Run("UI Redirection", func(t *testing.T) {
		redirectHandler, ok := mux.routes["/api/docs"]
		if !ok {
			t.Fatalf("Redirect handler not registered at /api/docs")
		}

		req := httptest.NewRequest("GET", "/api/docs", nil)
		w := httptest.NewRecorder()
		redirectHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMovedPermanently {
			t.Errorf("Expected status 301, got %d", resp.StatusCode)
		}
		location := resp.Header.Get("Location")
		if location != "/api/docs/" {
			t.Errorf("Expected Location /api/docs/, got %q", location)
		}
	})

	t.Run("Index serving", func(t *testing.T) {
		indexHandler, ok := mux.routes["/api/docs/"]
		if !ok {
			t.Fatalf("Index handler not registered at /api/docs/")
		}

		req := httptest.NewRequest("GET", "/api/docs/", nil)
		w := httptest.NewRecorder()
		indexHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "text/html") || !strings.Contains(ct, "charset=utf-8") {
			t.Errorf("Expected Content-Type text/html with charset=utf-8, got %q", ct)
		}
	})
}

func TestRegister_CustomAssetsPath(t *testing.T) {
	mux := &mockRouter{}
	l := zerolog.Nop()
	cfg := Config{
		UIPath:     "/api/docs/",
		JSONPath:   "/api/docs/swagger.json",
		FilePath:   "testdata/swagger.json",
		Host:       "localhost",
		AssetsPath: "static",
	}

	err := Register(mux, l, cfg)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	t.Run("Custom asset serving", func(t *testing.T) {
		assetHandler, ok := mux.routes["/api/docs/static/*"]
		if !ok {
			t.Fatalf("Asset handler not registered at /api/docs/static/*")
		}

		req := httptest.NewRequest("GET", "/api/docs/static/swagger-ui.css", nil)
		w := httptest.NewRecorder()
		assetHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Index serving with custom assets URL", func(t *testing.T) {
		indexHandler, ok := mux.routes["/api/docs/"]
		if !ok {
			t.Fatalf("Index handler not registered at /api/docs/")
		}

		req := httptest.NewRequest("GET", "/api/docs/", nil)
		w := httptest.NewRecorder()
		indexHandler(w, req)

		body := w.Body.String()
		if !strings.Contains(body, `href="static/swagger-ui.css"`) {
			t.Errorf("Expected body to contain custom asset path, but it didn't")
		}
	})
}
