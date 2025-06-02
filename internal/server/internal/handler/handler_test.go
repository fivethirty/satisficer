package handler_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/server/internal/handler"
)

type fakeBuilder struct {
	buildCount int
	err        error
}

func (b *fakeBuilder) Build(buildDir string) error {
	if b.err != nil {
		return b.err
	}
	b.buildCount++
	dest := filepath.Join(buildDir, "build_count.txt")
	return os.WriteFile(dest, []byte(strconv.Itoa(b.buildCount)), os.ModePerm)
}

type fakeWatcher struct {
	c <-chan time.Time
}

func newFakeWatcher(c chan time.Time) *fakeWatcher {
	return &fakeWatcher{
		c: c,
	}
}

func (w *fakeWatcher) C() <-chan time.Time {
	return w.c
}

func (w *fakeWatcher) Start() error {
	return nil
}
func (w *fakeWatcher) Stop() {
	// no-op
}

func TestHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "can start the server",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := make(chan time.Time)
			w := newFakeWatcher(c)
			b := &fakeBuilder{}
			_, err := handler.New(w, b)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
