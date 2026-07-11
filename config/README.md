### Config Package

The `config` package provides a centralized way to manage application configurations using environment variables and `.env` files.

#### Goal

The main objective of this package is to load, parse, and provide access to all configuration settings needed by the application, ensuring that sensitive data can be masked when logged.

#### How to Use

To initialize the configuration, call the `InitConfig` function at the start of your application:

```go
import "github.com/getjv/tbox/config"

func main() {
    cfg, err := config.InitConfig()
    if err != nil {
        // handle error
    }
    // use cfg
}
```

The configuration is loaded in the following order:
1.  Global `.env` file (if present).
2.  Environment-specific file `.env.{ENVIRONMENT}` (e.g., `.env.development`, `.env.production`).
3.  Direct environment variables (these take precedence via `envconfig`).

#### How to Extend

To add new configuration parameters:

1.  **Define a new struct** (if it's a new group of settings) or add fields to existing structs in the `config` package.
2.  **Add `envconfig` tags** to the struct fields to map them to environment variables.
3.  **Update `AppConfig` struct** in `config.go` to include your new settings.
4.  **Update `MarshalZerologObject`** in `config.go` if you want the new settings to be logged (and remember to mask sensitive data using `utils.MaskString`).

##### Local Data/Development

For local development, you can create a `.env.development` file in the project root. If you need to override settings without committing them to the repository, you can use the `ENV_PATH` environment variable to point to a different directory containing your private `.env` files, or simply set environment variables directly in your shell/IDE.
