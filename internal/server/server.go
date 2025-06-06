package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/fivethirty/satisficer/internal/builder"
	"github.com/fivethirty/satisficer/internal/server/internal/handler"
	"github.com/fivethirty/satisficer/internal/server/internal/watcher"
)

func Serve(projectFS fs.FS, port uint16) error {
	b, err := builder.New(projectFS)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ticker := time.NewTicker(1 * time.Second)
	w, err := watcher.Start(ctx, projectFS, ticker.C)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "satisficer-server-")
	if err != nil {
		return err
	}
	h, err := handler.Start(ctx, w, b, dir)
	if err != nil {
		return err
	}

	portStr := fmt.Sprintf(":%d", port)

	err = http.ListenAndServe(portStr, h)
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
