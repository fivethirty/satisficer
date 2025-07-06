package fsutil_test

import (
	"io/fs"
	"os"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/fivethirty/satisficer/internal/fsutil"
	"github.com/fivethirty/satisficer/internal/testutil"
)

func TestCopyFS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  fs.FS
	}{
		{
			name: "nil filesystem",
			src:  nil,
		},
		{
			name: "empty filesystem",
			src:  fstest.MapFS{},
		},
		{
			name: "filesystem with files",
			src: fstest.MapFS{
				"a/a.txt": &fstest.MapFile{},
				"a/b.txt": &fstest.MapFile{},
				"a.txt":   &fstest.MapFile{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			err := fsutil.CopyFS(test.src, dir)
			if err != nil {
				t.Fatal(err)
			}

			expected := testutil.SortedPaths(t, test.src)
			sort.Strings(expected)

			actual := testutil.SortedPaths(t, os.DirFS(dir))
			sort.Strings(actual)

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("expected paths %v, got %v", expected, actual)
			}
		})
	}
}
