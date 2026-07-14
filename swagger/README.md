# Swagger

This package provides a simple way to register and serve Swagger UI for your Go application using [swaggo/swag](https://github.com/swaggo/swag).

## Objective

The goal of this package is to automate the process of serving OpenAPI documentation generated from your code's comments. It embeds both the Swagger UI (HTML/JS/CSS assets) and the OpenAPI specification to ensure a self-contained and secure documentation environment, without relying on external CDNs.

## Security

By embedding the Swagger UI assets locally, this package avoids the security risks associated with third-party CDNs (like [unpkg.com](https://unpkg.com)), such as potential script injection or service unavailability. It also ensures that the documentation remains accessible even in environments with restricted internet access.

Additionally, all embedded assets (CSS and JS) are served with **Subresource Integrity (SRI)** attributes (`integrity` and `crossorigin`) using SHA-512 hashes. The attributes are placed before the source path, and relative paths are kept clean (avoiding `./` prefixes) to ensure maximum compatibility with security scanners like SonarQube.

## Installation

To generate documentation, you must have the `swag` CLI tool installed. You can install it locally or use it in a CI/CD pipeline like GitHub Actions.

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Initial Configuration

To start using Swagger, add the general API annotations to your `main.go` file:

```go
package main

// @title YOUR_APP_NAME
// @version 1.0.0
// @description YOUR_APP_DESCRIPTION

// @BasePath /

// @securityDefinitions.oauth2.password AuthForm
// @tokenUrl /api/auth/login

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and then your token.

func main() {
    // ...
}
```

## How to Use

### Registering with the Router

This package is router-agnostic. It defines a `Router` interface that expects a `Get` method. Most popular routers like `chi`, `gin`, and `echo` implement this method and can be used directly.

#### Using with Chi (or similar routers)

Register it in your server setup:

```go
// Register Swagger
err := swagger.Register(s.mux, s.logger, swagger.Config{
    UIPath:   "/api/docs/",               // URL path for the UI
    JSONPath: "/api/docs/swagger.json",    // URL path for the JSON spec
    FilePath: "backend/docs/swagger.json", // Local path to the generated JSON file
    Host:     s.cfg.Web.Address(),        // Host address for logging purposes
    AssetsPath: "assets",                  // Optional: URL path segment for assets (defaults to "assets")
})
if err != nil {
    s.logger.Error().Err(err).Msg("Failed to register swagger")
}
```

#### Using with standard http.ServeMux

Since `http.ServeMux` does not have a `Get` method, you can use a simple adapter:

```go
type muxAdapter struct {
    *http.ServeMux
}

func (m muxAdapter) Get(path string, h http.HandlerFunc) {
    m.HandleFunc("GET " + path, h)
}

// ... in your setup
mux := http.NewServeMux()
err := swagger.Register(muxAdapter{mux}, logger, cfg)
```

## How to Extend

Add annotations to your handler functions to document specific endpoints.

```go
// @Summary Get book details
// @Description Returns the full thread of a book, including chapters, exercises, and questions
// @Tags books
// @Produce json
// @Security AuthForm
// @Security BearerAuth
// @Param bookID path string true "Book ID"
// @Success 200 {object} contracts.BookThread
// @Failure 401 {object} contracts.ErrorResponse
// @Failure 404 {object} contracts.ErrorResponse
// @Failure 500 {object} contracts.ErrorResponse
// @Router /api/books/{bookID}/details [get]
func (h *BookHandler) GetBookDetails(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

> **Note:** For authentication to work in Swagger UI, ensure the `@Security` annotations match the definitions in your `main.go`.

## Generating Documentation

You can use a `Makefile` command to automate the generation of the Swagger documentation. Here is an example of the command and an explanation of its parameters:

```makefile
docs:
	swag init -g backend/cmd/api/main.go -o backend/docs --ot json --parseDependency --parseInternal
	rm -f backend/docs/swagger.yaml backend/docs/docs.go
```

### Explanation of Parameters:
- `-g backend/cmd/api/main.go`: Specifies the "general" API information file (where your `@title`, `@version`, etc., are located).
- `-o backend/docs`: Defines the output directory for the generated specification files.
- `--ot json`: Ensures that only the JSON format is generated (OpenAPI 2.0/3.0 JSON).
- `--parseDependency`: Tells `swag` to parse outside dependencies to find models used in your API.
- `--parseInternal`: Allows `swag` to parse internal packages within your project.

### Why remove `swagger.yaml` and `docs.go`?
- **`swagger.yaml`**: Since this package is configured to serve the `.json` version of the spec, the YAML version is redundant and is removed to keep the directory clean.
- **`docs.go`**: By default, `swag` generates a `docs.go` file that embeds the entire specification as a string. However, this package is designed to serve the `swagger.json` file directly from the filesystem (or you could embed it yourself). Removing `docs.go` avoids unnecessary code generation and keeps the binary size smaller.

> **Important:** The paths in the example above (`backend/cmd/api/main.go` and `backend/docs`) should be adjusted to match your actual project structure.
