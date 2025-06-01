package logs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type Handler struct {
	slog.Handler
	mu *sync.Mutex
	w  io.Writer
}

func NewHandler(w io.Writer) *Handler {
	if w == nil {
		w = os.Stderr
	}
	return &Handler{
		Handler: slog.NewTextHandler(w, &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			AddSource:   false,
			ReplaceAttr: nil,
		}),
		mu: &sync.Mutex{},
		w:  w,
	}
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	strs := []string{
		fmt.Sprintf("[%s]", r.Level.String()),
		r.Message,
	}

	if r.NumAttrs() != 0 {
		r.Attrs(func(a slog.Attr) bool {
			attr := fmt.Sprintf("%s=%s", a.Key, a.Value.String())
			strs = append(strs, attr)
			return true
		})
	}

	result := strings.Join(strs, " ")
	b := []byte(result)
	b = append(b, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.w.Write(b)

	return err
}
