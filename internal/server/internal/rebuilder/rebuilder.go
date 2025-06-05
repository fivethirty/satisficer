package rebuilder

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

type Watcher interface {
	Ch() <-chan time.Time
}

type Builder interface {
	Build(buildDir string) error
}

type Rebuilder struct {
	BuildDir string
	watcher  Watcher
	builder  Builder
	atomicFS atomic.Value
	ch       chan error
	baseDir  string
}

func New(w Watcher, b Builder, baseDir string) Rebuilder {
	return Rebuilder{
		watcher: w,
		builder: b,
		ch:      make(chan error, 1),
		baseDir: baseDir,
	}
}

func Start(ctx context.Context, w Watcher, b Builder, baseDir string) (*Rebuilder, error) {
	r := Rebuilder{
		watcher: w,
		builder: b,
		ch:      make(chan error, 1),
		baseDir: baseDir,
	}
	r.publish()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.watcher.Ch():
				r.publish()
			}
		}
	}()
	return &r, nil
}

func (r *Rebuilder) publish() {
	select {
	case r.ch <- r.rebuild():
	default:
	}
}

func (r *Rebuilder) Ch() <-chan error {
	return r.ch
}

func (r *Rebuilder) Latest() fs.FS {
	return r.atomicFS.Load().(fs.FS)
}

func (r *Rebuilder) rebuild() error {
	dir, err := os.MkdirTemp(r.baseDir, "satisficer-server-build-")
	if err != nil {
		return err
	}

	toRemove := dir
	buildErr := r.builder.Build(dir)
	if buildErr == nil {
		r.atomicFS.Store(os.DirFS(dir))
		toRemove = r.BuildDir
		r.BuildDir = dir
	}

	if err := os.RemoveAll(toRemove); err != nil {
		slog.Warn("failed to remove build directory", "path", toRemove, "error", err)
	}

	return buildErr
}
