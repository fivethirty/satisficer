package rebuilder

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Watcher interface {
	Start() error
	Stop()
	ChangedCh() <-chan time.Time
}

type Builder interface {
	Build(buildDir string) error
}

type state int

const (
	created state = iota
	running
	stopped
)

type Rebuilder struct {
	watcher    Watcher
	builder    Builder
	atomicFS   atomic.Value
	doneCh     chan bool
	rebuiltCh  chan error
	state      state
	stateMutex sync.Mutex
	baseDir    string
	buildDir   string
}

func New(w Watcher, b Builder, baseDir string) Rebuilder {
	return Rebuilder{
		watcher:    w,
		builder:    b,
		doneCh:     make(chan bool),
		rebuiltCh:  make(chan error, 1),
		state:      created,
		stateMutex: sync.Mutex{},
		baseDir:    baseDir,
	}
}

func (r *Rebuilder) Start() error {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	if r.state != created {
		return fmt.Errorf("rebuilder has already been started")
	}
	r.state = running
	if err := r.watcher.Start(); err != nil {
		return fmt.Errorf("could not start watcher: %w", err)
	}
	r.rebuiltCh <- r.rebuild()
	go r.watch()
	return nil
}

func (r *Rebuilder) Stop() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	if r.state != running {
		return
	}
	r.state = stopped
	r.doneCh <- true
	close(r.doneCh)
	r.watcher.Stop()

	if err := os.RemoveAll(r.baseDir); err != nil {
		slog.Warn("failed to remove base rebuilder directory", "path", r.baseDir, "error", err)
	}
}

func (r *Rebuilder) RebuiltCh() <-chan error {
	return r.rebuiltCh
}

func (r *Rebuilder) FS() fs.FS {
	return r.atomicFS.Load().(fs.FS)
}

func (r *Rebuilder) watch() {
	for {
		select {
		case <-r.doneCh:
			return
		case <-r.watcher.ChangedCh():
			r.rebuiltCh <- r.rebuild()
		}
	}
}

func (r *Rebuilder) rebuild() error {
	dir, err := os.MkdirTemp(r.baseDir, "satisficer-server-build-")
	if err != nil {
		return err
	}

	resultErr := r.builder.Build(dir)

	if err := os.RemoveAll(r.buildDir); err != nil {
		slog.Warn("failed to remove old build directory", "path", r.buildDir, "error", err)
	}

	r.buildDir = dir
	return resultErr
}
