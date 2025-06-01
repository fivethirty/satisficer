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
	mu  *sync.Mutex
	out io.Writer
}

func NewHandler(out io.Writer) *Handler {
	if out == nil {
		out = os.Stderr
	}
	return &Handler{
		Handler: slog.NewTextHandler(out, &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			AddSource:   false,
			ReplaceAttr: nil,
		}),
		mu:  &sync.Mutex{},
		out: out,
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

	_, err := h.out.Write(b)

	return err

}
