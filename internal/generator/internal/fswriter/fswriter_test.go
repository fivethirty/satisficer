package fswriter_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/fivethirty/satisficer/internal/generator/internal/fswriter"
)

func TestCopyFilteredFS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		srcFS          fs.FS
		filter         fswriter.PathFilterFunc
		expectedDestFS fs.FS
	}{
		{
			name: "can copy all files",
			srcFS: fstest.MapFS{
				"file1.txt": &fstest.MapFile{
					Data: []byte("file1 content"),
				},
				"dir/file2.txt": &fstest.MapFile{
					Data: []byte("file2 content"),
				},
			},
			filter: fswriter.AllPathFilterFunc,
			expectedDestFS: fstest.MapFS{
				"file1.txt": &fstest.MapFile{
					Data: []byte("file1 content"),
				},
				"dir/file2.txt": &fstest.MapFile{
					Data: []byte("file2 content"),
				},
			},
		},
		{
			name: "can filter out specific files",
			srcFS: fstest.MapFS{
				"file1.md": &fstest.MapFile{
					Data: []byte("# file1"),
				},
				"file2.txt": &fstest.MapFile{
					Data: []byte("file2 content"),
				},
			},
			filter: func(path string) (bool, error) {
				return filepath.Ext(path) != ".md", nil
			},
			expectedDestFS: fstest.MapFS{
				"file2.txt": &fstest.MapFile{
					Data: []byte("file2 content"),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			err := fswriter.CopyFilteredFS(test.srcFS, dir, test.filter)
			if err != nil {
				t.Fatal(err)
			}
			actualDestFS := toMapFS(t, os.DirFS(dir))
			if !reflect.DeepEqual(actualDestFS, toMapFS(t, test.expectedDestFS)) {
				t.Errorf("expected %v, got %v", toMapFS(t, test.expectedDestFS), actualDestFS)
			}
		})
	}
}

func toMapFS(t *testing.T, fsys fs.FS) fstest.MapFS {
	t.Helper()
	mfs := fstest.MapFS{}
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		mfs[path] = &fstest.MapFile{
			Data: data,
			Mode: d.Type(),
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return mfs
}
