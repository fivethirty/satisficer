package fswriter

import (
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

func Copy(src io.Reader, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, src); err != nil {
		return err
	}
	return nil
}

type PathFilterFunc func(path string) (bool, error)

func AllPathFilterFunc(path string) (bool, error) {
	return true, nil
}

func CopyFilteredFS(src fs.FS, dest string, filter PathFilterFunc) error {
	err := fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ok, err := filter(path)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		slog.Info("Copying file", "src", path, "dest", filepath.Join(dest, path))
		src, err := src.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = src.Close() }()
		destPath := filepath.Join(dest, path)
		return Copy(src, destPath)
	})
	if err != nil {
		return err
	}
	return nil
}
