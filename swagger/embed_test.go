package swagger

import (
	"fmt"
	"io/fs"
	"testing"
)

func TestEmbedFS(t *testing.T) {
	err := fs.WalkDir(assets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Printf("path: %s, isDir: %v\n", path, d.IsDir())
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk assets: %v", err)
	}
}
