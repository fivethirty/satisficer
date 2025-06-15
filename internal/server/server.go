package server

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	ticker := time.NewTicker(300 * time.Millisecond)
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	slog.Info(
		"Starting server",
		"port", portStr,
	)

	serverErr := make(chan error, 1)
	go func() {
		if err := http.ListenAndServe(portStr, h); err != nil {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-sigChan:
		if err := os.RemoveAll(dir); err != nil {
			slog.Warn("failed to remove server build directory", "path", dir, "error", err)
		}
		return nil
	}
}
