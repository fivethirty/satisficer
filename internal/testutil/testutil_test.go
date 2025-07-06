package testutil_test

import (
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/fivethirty/satisficer/internal/testutil"
)

func TestSortedPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    fs.FS
		expected []string
	}{
		{
			name:     "empty filesystem",
			input:    nil,
			expected: []string{},
		},
		{
			name: "single file",
			input: fstest.MapFS{
				"testdata/single.txt": &fstest.MapFile{},
			},
			expected: []string{"testdata/single.txt"},
		},
		{
			name: "many file",
			input: fstest.MapFS{
				"a/a.txt": &fstest.MapFile{},
				"a/b.txt": &fstest.MapFile{},
				"a.txt":   &fstest.MapFile{},
				"b.txt":   &fstest.MapFile{},
			},
			expected: []string{
				"a.txt",
				"a/a.txt",
				"a/b.txt",
				"b.txt",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := testutil.SortedPaths(t, test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}
