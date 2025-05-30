package creator_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/fivethirty/satisficer/internal/creator"
	"github.com/fivethirty/satisficer/internal/testutil"
)

func TestCreate(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "satisficer_creator-TestCreate")
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	err := creator.Create(dir)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	expectedPaths := []string{
		"content/index.md",
		"layout/index.html.tmpl",
		"layout/static/main.css",
	}
	sort.Strings(expectedPaths)

	actualPaths := testutil.SortedPaths(t, os.DirFS(dir))
	sort.Strings(actualPaths)

	if !reflect.DeepEqual(actualPaths, expectedPaths) {
		t.Errorf("expected paths %v, got %v", expectedPaths, actualPaths)
	}
}

func TestCreateError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	err := creator.Create(dir)
	if err == nil {
		t.Fatal("expected an error when creating in an existing directory, got nil")
	}
}
