// Package openai provides a reusable OpenAI chat-completions client with retry logic.
// Domain-specific GPT clients should embed or compose this client instead of duplicating
// the HTTP transport, retry, and JSON decoding logic.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// ChatMessage represents a single message in a chat-completions request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the payload sent to the OpenAI chat-completions endpoint.
type ChatRequest struct {
	Model          string          `json:"model"`
	Messages       []ChatMessage   `json:"messages"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// ResponseFormat controls the output format of the chat-completions response.
type ResponseFormat struct {
	Type string `json:"type"`
}

// ChatResponse is the payload returned by the OpenAI chat-completions endpoint.
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// Config provides access to OpenAI configuration.
type Config interface {
	GetOpenAIAPIKey() string
	GetOpenAIBaseURL() string
	GetOpenAIModel() string
}

// WithTimeout sets the HTTP client timeout. Default is 2 minutes.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithMaxRetries sets the maximum number of retry attempts. Default is 3.
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// WithBackoffMultiplier sets the base duration multiplied by the retry index
// to compute the wait time between retries. Default is 1 second.
func WithBackoffMultiplier(d time.Duration) ClientOption {
	return func(c *Client) {
		c.backoffMultiplier = d
	}
}

// Client is a reusable OpenAI chat-completions HTTP client.
type Client struct {
	apiKey            string
	baseURL           string
	model             string
	logger            zerolog.Logger
	httpClient        *http.Client
	maxRetries        int
	backoffMultiplier time.Duration
}

// NewClient creates a new OpenAI chat-completions client.
func NewClient(cfg Config, logger zerolog.Logger, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:  cfg.GetOpenAIAPIKey(),
		baseURL: cfg.GetOpenAIBaseURL(),
		model:   cfg.GetOpenAIModel(),
		logger:  logger,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		maxRetries:        3,
		backoffMultiplier: 1 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Model returns the model name used by this client.
func (c *Client) Model() string {
	return c.model
}

// Do sends a chat-completions request and returns the raw content string from the first choice.
// It handles retries with exponential backoff for transient (5xx) errors.
func (c *Client) Do(ctx context.Context, req ChatRequest) (string, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	for _, msg := range req.Messages {
		c.logger.Info().Str("role", msg.Role).Str("content", msg.Content).Msg("GPT Prompt")
	}

	var lastErr error
	for i := range c.maxRetries {
		if err := c.waitBeforeRetry(ctx, i); err != nil {
			return "", err
		}

		content, shouldRetry, err := c.doRequest(ctx, jsonData, i)
		if err != nil {
			if shouldRetry {
				lastErr = err
				continue
			}
			return "", err
		}
		return content, nil
	}

	return "", fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *Client) waitBeforeRetry(ctx context.Context, attempt int) error {
	if attempt == 0 {
		return nil
	}
	c.logger.Info().Int("retry", attempt).Msg("Retrying GPT request after error")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(attempt) * c.backoffMultiplier):
		return nil
	}
}

func (c *Client) doRequest(ctx context.Context, jsonData []byte, attempt int) (string, bool, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", false, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Warn().Err(err).Int("attempt", attempt+1).Msg("GPT request failed, might retry")
		return "", true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode >= 500, c.handleErrorResponse(resp)
	}

	return c.decodeResponse(resp.Body)
}

func (c *Client) handleErrorResponse(resp *http.Response) error {
	respBody, _ := io.ReadAll(resp.Body)
	err := fmt.Errorf("GPT API returned error status: %d", resp.StatusCode)
	c.logger.Error().Int("status", resp.StatusCode).Str("body", string(respBody)).Msg("GPT API error")
	return err
}

func (c *Client) decodeResponse(body io.Reader) (string, bool, error) {
	var chatResp ChatResponse
	if err := json.NewDecoder(body).Decode(&chatResp); err != nil {
		return "", true, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", true, fmt.Errorf("no choices returned from GPT")
	}

	content := chatResp.Choices[0].Message.Content
	c.logger.Info().Str("content", content).Msg("GPT Response")
	return content, false, nil
}
