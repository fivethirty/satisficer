package handler

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fivethirty/satisficer/internal/server/internal/handler/responsebody"
)

type Watcher interface {
	Ch() <-chan time.Time
}

type Builder interface {
	Build(buildDir string) error
}

type build struct {
	dir        string
	fileServer http.Handler
	err        error
}

type Handler struct {
	watcher Watcher
	builder Builder
	baseDir string
	build   atomic.Value
	buildCh chan time.Time
	mux     *http.ServeMux
	ctx     context.Context
}

func Start(ctx context.Context, w Watcher, b Builder, baseDir string) (*Handler, error) {
	h := Handler{
		watcher: w,
		builder: b,
		buildCh: make(chan time.Time, 1),
		baseDir: baseDir,
		ctx:     ctx,
	}

	h.rebuild()
	h.watch()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /_/events", h.events)
	mux.HandleFunc("GET /", h.files)
	h.mux = mux

	return &h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) watch() {
	go func() {
		for {
			select {
			case <-h.ctx.Done():
				return
			case <-h.watcher.Ch():
				h.rebuild()
				h.publish()
			}
		}
	}()
}

func (h *Handler) events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	for {
		select {
		case <-h.buildCh:
			_, err := fmt.Fprint(w, "data: rebuild\n\n")
			if err != nil {
				http.Error(w, "failed to write event", http.StatusInternalServerError)
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-h.ctx.Done():
			return
		}
	}
}

type bufResponseWriter struct {
	buf        bytes.Buffer
	statusCode int
	header     http.Header
}

func (b *bufResponseWriter) Write(data []byte) (int, error) {
	return b.buf.Write(data)
}

func (b *bufResponseWriter) WriteHeader(statusCode int) {
	b.statusCode = statusCode
}

func (b *bufResponseWriter) Header() http.Header {
	if b.header == nil {
		b.header = make(http.Header)
	}
	return b.header
}

func (h *Handler) files(w http.ResponseWriter, r *http.Request) {
	build, ok := h.build.Load().(build)
	if !ok {
		http.Error(w, "not built yet", http.StatusInternalServerError)
		return
	}
	if build.err != nil {
		http.Error(w, build.err.Error(), http.StatusInternalServerError)
		return
	}

	wrapped := &bufResponseWriter{
		buf: bytes.Buffer{},
	}

	build.fileServer.ServeHTTP(wrapped, r)

	if wrapped.statusCode >= 300 && wrapped.statusCode < 400 {
		for key, values := range wrapped.header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(wrapped.statusCode)
		return
	}

	bufContent := wrapped.buf.Bytes()

	contentType := wrapped.header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		bufContent = responsebody.WithReloadHTML(bufContent)
	}

	for key, values := range wrapped.header {
		if key == "Content-Length" || key == "Last-Modified" || key == "Etag" {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	_, err := w.Write(bufContent)
	if err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) publish() {
	select {
	case h.buildCh <- time.Now():
	default:
	}
}

func (h *Handler) rebuild() {
	oldBuild, ok := h.build.Load().(build)
	if ok {
		oldDir := oldBuild.dir
		defer func() {
			if oldDir == "" {
				return
			}
			if err := os.RemoveAll(oldDir); err != nil {
				slog.Warn("failed to remove build directory", "path", oldDir, "error", err)
			}
		}()
	}

	dir, err := os.MkdirTemp(h.baseDir, "satisficer-server-build-")
	if err != nil {
		h.build.Store(build{err: err})
		return
	}

	buildErr := h.builder.Build(dir)
	if buildErr != nil {
		slog.Error("build failed", "error", buildErr)
	}

	h.build.Store(
		build{
			dir:        dir,
			err:        buildErr,
			fileServer: http.FileServer(http.Dir(dir)),
		},
	)
}
