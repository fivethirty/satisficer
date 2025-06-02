package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type Handler struct {
	atomicSuccessHandler atomic.Value
	currentError         error
	watcher              Watcher
	builder              Builder
}

type Watcher interface {
	Start() error
	Stop()
	C() <-chan time.Time
}

type Builder interface {
	Build(buildDir string) error
}

func New(w Watcher, b Builder) (*Handler, error) {
	h := &Handler{
		atomicSuccessHandler: atomic.Value{},
		watcher:              w,
		builder:              b,
	}
	buildDir, err := h.build()
	if err != nil {
		return nil, fmt.Errorf("could not build project: %w", err)
	}
	if err := h.watcher.Start(); err != nil {
		return nil, fmt.Errorf("could not start watcher: %w", err)
	}
	go h.watch(buildDir)
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

func (h *Handler) watch(prevBuildDir string) {
	// defer is not guarenteed to happen here need to think of something better
	defer func() {
		fmt.Println("Ending watcher...")
		h.watcher.Stop()
		_ = os.RemoveAll(prevBuildDir)
	}()

	for t := range h.watcher.C() {
		slog.Info("Change detected", "time", t)
		buildDir, err := h.build()
		if err != nil {
			slog.Error("Build failed", "error", err)
			h.currentError = fmt.Errorf("build failed: %w", err)
			continue
		}
		successHandler := http.StripPrefix(buildDir, http.FileServer(http.Dir(buildDir)))
		h.atomicSuccessHandler.Store(successHandler)
		h.currentError = nil
		if err := os.RemoveAll(prevBuildDir); err != nil {
			slog.Warn("Cold not remove previous build directory", "error", err)
		}
		prevBuildDir = buildDir
	}
}

func (h *Handler) build() (string, error) {
	buildDir, err := os.MkdirTemp("", "satisficer-server-build-")
	if err != nil {
		return "", fmt.Errorf("error creating build directory: %w", err)
	}
	if err := h.builder.Build(buildDir); err != nil {
		_ = os.RemoveAll(buildDir)
		return "", fmt.Errorf("error building site: %w", err)
	}
	return buildDir, nil
}
