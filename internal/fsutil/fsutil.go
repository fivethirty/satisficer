package fsutil

import (
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

func CopyFS(src fs.FS, destDir string) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return CopyFile(src, path, destDir)
	})
}

func CopyFile(fsys fs.FS, path string, destDir string) error {
	slog.Info("Writing file", "path", path)
	src, err := fsys.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()
	dest, err := CreateFile(filepath.Join(destDir, path))
	if err != nil {
		return err
	}
	defer func() { _ = dest.Close() }()
	_, err = io.Copy(dest, src)
	return err
}

func CreateFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}

	destFile, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return destFile, nil
}
