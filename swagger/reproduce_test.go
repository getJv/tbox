package swagger

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestRegister_Reproduce(t *testing.T) {
	mux := &mockRouter{}
	l := zerolog.Nop()
	cfg := Config{
		UIPath:   "/api/docs/",
		JSONPath: "/api/docs/swagger.json",
		FilePath: "testdata/swagger.json",
	}

	err := Register(mux, l, cfg)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	assetHandler, ok := mux.routes["/api/docs/assets/*"]
	if !ok {
		t.Fatalf("Asset handler not registered at /api/docs/assets/*")
	}

	t.Run("Serve CSS", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/docs/assets/swagger-ui.css", nil)
		w := httptest.NewRecorder()
		assetHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "text/css") {
			t.Errorf("Expected Content-Type text/css, got %q", ct)
		}
	})

	t.Run("Serve JS Bundle", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/docs/assets/swagger-ui-bundle.js", nil)
		w := httptest.NewRecorder()
		assetHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "javascript") {
			t.Errorf("Expected Content-Type javascript, got %q", ct)
		}
	})

	t.Run("Redirect no-slash UIPath", func(t *testing.T) {
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

	t.Run("Serve Index at UIPath with slash", func(t *testing.T) {
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
		if !strings.Contains(ct, "text/html") {
			t.Errorf("Expected Content-Type text/html, got %q", ct)
		}
	})
}
