package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Handler struct {
	watcher              Watcher
	builder              Builder
	done                 chan bool
	isRunning            bool
	stopMutex            sync.Mutex
	atomicSuccessHandler atomic.Value
	buildDirBase         string
	activeBuildDir       string
	buildDirs            map[string]struct{}
	currentError         error
}

type Watcher interface {
	Start() error
	Stop()
	C() <-chan time.Time
}

type Builder interface {
	Build(buildDir string) error
}

func New(w Watcher, b Builder, buildDirBase string) (*Handler, error) {
	h := &Handler{
		watcher:              w,
		builder:              b,
		done:                 make(chan bool),
		isRunning:            true,
		stopMutex:            sync.Mutex{},
		atomicSuccessHandler: atomic.Value{},
		buildDirBase:         buildDirBase,
		activeBuildDir:       "",
		buildDirs:            make(map[string]struct{}),
		currentError:         nil,
	}
	h.rotateSuccessHandler()
	if err := h.watcher.Start(); err != nil {
		return nil, fmt.Errorf("could not start watcher: %w", err)
	}
	go h.watch()
	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.currentError != nil {
		msg := fmt.Sprintf(
			"Error building site, please see terminal logs for more info: %v",
			h.currentError,
		)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	sh, ok := h.atomicSuccessHandler.Load().(http.Handler)
	if !ok {
		http.Error(w, "Internal server error: handler not set", http.StatusInternalServerError)
		return
	}

	sh.ServeHTTP(w, r)
}

func (h *Handler) Stop() {
	h.stopMutex.Lock()
	defer h.stopMutex.Unlock()
	if !h.isRunning {
		return
	}
	h.isRunning = false
	h.done <- true
	close(h.done)
}

func (h *Handler) watch() {
	for {
		select {
		case <-h.done:
			return
		case t := <-h.watcher.C():
			slog.Info("Change detected", "time", t)
			h.rotateSuccessHandler()
		}
	}
}

func (h *Handler) rotateSuccessHandler() {
	if err := h.buildAndCleanup(); err != nil {
		slog.Error("Build failed", "error", err)
		h.currentError = err
		return
	}
	h.atomicSuccessHandler.Store(http.FileServer(http.Dir(h.activeBuildDir)))
	h.currentError = nil
}

func (h *Handler) buildAndCleanup() error {
	buildDir, err := os.MkdirTemp(h.buildDirBase, "satisficer-server-build-")
	if err != nil {
		return fmt.Errorf("error creating build directory: %w", err)
	}
	h.activeBuildDir = buildDir
	h.buildDirs[buildDir] = struct{}{}

	for dir := range h.buildDirs {
		if dir == buildDir {
			continue
		}
		if err := os.RemoveAll(dir); err != nil {
			slog.Warn("Could not remove old build directory", "dir", dir, "error", err)
		} else {
			delete(h.buildDirs, dir)
		}
	}

	if err := h.builder.Build(buildDir); err != nil {
		return fmt.Errorf("error building site: %w", err)
	}
	return nil
}
