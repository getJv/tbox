# OpenAI Client

A reusable Go client for OpenAI's chat-completions API with built-in retry logic and exponential backoff.

## Features

- **Retry Logic**: Automatically retries on 5xx errors with exponential backoff.
- **Configurable**: Easily configured via an interface or functional options.
- **Logging**: Integrated with `zerolog` for transparent operation.
- **JSON Support**: Supports `response_format` for structured outputs (e.g., JSON mode).

## Installation

```bash
go get github.com/getjv/tbox/openai
```

## Configuration

The `NewClient` function requires a `Config` interface:

```go
type Config interface {
    GetOpenAIAPIKey() string
    GetOpenAIBaseURL() string
    GetOpenAIModel() string
}
```

### Environment Variables

If you are using `envconfig` or similar, you can use the following environment variables:

- `OPENAI_API_KEY`: Your OpenAI API key.
- `OPENAI_BASE_URL`: The API endpoint (defaults to `https://api.openai.com/v1/chat/completions`).
- `OPENAI_MODEL`: The model to use (defaults to `gpt-4o-mini`).

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/getjv/tbox/openai"
    "github.com/rs/zerolog"
    "os"
)

type myConfig struct{}
func (c myConfig) GetOpenAIAPIKey() string { return os.Getenv("OPENAI_API_KEY") }
func (c myConfig) GetOpenAIBaseURL() string { return "https://api.openai.com/v1/chat/completions" }
func (c myConfig) GetOpenAIModel() string { return "gpt-4o-mini" }

func main() {
    logger := zerolog.New(os.Stdout)
    cfg := myConfig{}
    
    client := openai.NewClient(cfg, logger)
    
    ctx := context.Background()
    req := openai.ChatRequest{
        Messages: []openai.ChatMessage{
            {Role: "user", Content: "Hello, how are you?"},
        },
    }
    
    response, err := client.Do(ctx, req)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(response)
}
```

### Using ClientOption

You can further customize the client using `ClientOption`:

```go
client := openai.NewClient(cfg, logger, 
    openai.WithTimeout(30 * time.Second),
    openai.WithMaxRetries(5),
    openai.WithBackoffMultiplier(500 * time.Millisecond),
)
```

- `WithTimeout(d time.Duration)`: Sets the HTTP client timeout (default: 2 minutes).
- `WithMaxRetries(n int)`: Sets the maximum number of retry attempts (default: 3).
- `WithBackoffMultiplier(d time.Duration)`: Sets the base duration for exponential backoff (default: 1 second).

