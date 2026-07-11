### Utils Package

The `utils` package provides a collection of generic and multi-use utility functions for common tasks like string manipulation and HTTP responses.

#### Goal

The main objective of this package is to avoid code duplication by providing reusable functions that can be used throughout the entire project.

#### How to Use

##### String Utilities

```go
import "github.com/getjv/tbox/utils"

masked := utils.MaskString("sensitive-data") // returns "******" or "abcde...abcde"
sentence := utils.MaskRandomWord("Hello world") // returns e.g. "Hello ____"
```

##### HTTP Response Utilities

```go
import "github.com/getjv/tbox/utils"

func MyHandler(w http.ResponseWriter, r *http.Request) {
    utils.RespondWithJSON(w, http.StatusOK, myData)
    // or
    utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
}
```

#### How to Extend

To add new utilities:

1.  **Create a new file** in the `utils` directory if it's a new category of utilities (e.g., `math.go`).
2.  **Add your functions** to the `utils` package.
3.  **Add tests** in a corresponding `*_test.go` file to ensure the new utility works as expected.

For local/private utilities that are only relevant to a specific module, consider keeping them within that module instead of adding them here. This package is intended for functions that are truly generic and widely useful across the repository.
