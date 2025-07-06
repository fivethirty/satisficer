package creator

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/fivethirty/satisficer/internal/fsutil"
)

const (
	dirPerm = 0750
)

//go:embed starter
var starter embed.FS

func Create(dir string) error {
	slog.Info("Creating project", "dir", dir)
	_, err := os.Stat(dir)
	if err == nil {
		return fmt.Errorf("'%s' already exists", dir)
	}

	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return err
	}

	subFS, err := fs.Sub(starter, "starter")
	if err != nil {
		return err
	}

	err = fsutil.CopyFS(subFS, dir)
	if err != nil {
		return err
	}
	slog.Info("Project created successfully", "dir", dir)
	return nil
}
