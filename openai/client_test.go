package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfig struct {
	apiKey  string
	baseURL string
	model   string
}

func (m mockConfig) GetOpenAIAPIKey() string  { return m.apiKey }
func (m mockConfig) GetOpenAIBaseURL() string { return m.baseURL }
func (m mockConfig) GetOpenAIModel() string   { return m.model }

func TestClient_Do_Success(t *testing.T) {
	logger := zerolog.Nop()
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-key")

		resp := ChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `{"result":"ok"}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := mockConfig{apiKey: "test-key", baseURL: server.URL, model: "gpt-4o-mini"}
	client := NewClient(cfg, logger)

	result, err := client.Do(ctx, ChatRequest{
		Messages:       []ChatMessage{{Role: "user", Content: "hello"}},
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	})

	require.NoError(t, err)
	assert.Equal(t, `{"result":"ok"}`, result)
}

func TestClient_Do_Retry_On_5xx(t *testing.T) {
	logger := zerolog.Nop()
	ctx := context.Background()
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"server error"}`))
			return
		}

		resp := ChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "success"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := mockConfig{apiKey: "test-key", baseURL: server.URL, model: "gpt-4o-mini"}
	client := NewClient(cfg, logger, WithBackoffMultiplier(10*time.Millisecond))

	result, err := client.Do(ctx, ChatRequest{
		Messages: []ChatMessage{{Role: "user", Content: "hello"}},
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, attempts)
}

func TestClient_Do_NoRetry_On_4xx(t *testing.T) {
	logger := zerolog.Nop()
	ctx := context.Background()
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	cfg := mockConfig{apiKey: "test-key", baseURL: server.URL, model: "gpt-4o-mini"}
	client := NewClient(cfg, logger)

	_, err := client.Do(ctx, ChatRequest{
		Messages: []ChatMessage{{Role: "user", Content: "hello"}},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Equal(t, 1, attempts)
}
