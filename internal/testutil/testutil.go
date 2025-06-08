package testutil

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"testing"
)

func Ptr[T any](t *testing.T, tt T) *T {
	t.Helper()
	return &tt
}

func ToContent(t *testing.T, fmMap map[string]any, markdown string) string {
	t.Helper()
	frontMatter, err := json.Marshal(fmMap)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("---\n%s\n---\n%s", frontMatter, markdown)
}

func SortedPaths(t *testing.T, fsys fs.FS) []string {
	paths := []string{}
	if fsys == nil {
		return paths
	}
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	return paths
}
