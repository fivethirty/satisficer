package rebuilder_test

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/server/internal/rebuilder"
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

func (w *fakeWatcher) ChangedCh() <-chan time.Time {
	return w.c
}

func (w *fakeWatcher) Start() error {
	return nil
}
func (w *fakeWatcher) Stop() {
	// no-op
}

func TestRebuilder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		actions        func(trigger chan<- time.Time, builder fakeBuilder)
		wantBuildCount int
	}{
		{
			name: "initial build happens on start",
			actions: func(trigger chan<- time.Time, builder fakeBuilder) {
				// no op
			},
			wantBuildCount: 1,
		},
		{
			name: "file change triggers rebuild",
			actions: func(trigger chan<- time.Time, builder fakeBuilder) {
				trigger <- time.Now()
				time.Sleep(100 * time.Millisecond)
			},
			wantBuildCount: 2,
		},
		{
			name: "multiple error does not count as a build",
			actions: func(trigger chan<- time.Time, builder fakeBuilder) {
				trigger <- time.Now()
				time.Sleep(100 * time.Millisecond)
				builder.err = errors.New("fake error")
				trigger <- time.Now()
				time.Sleep(100 * time.Millisecond)
				builder.err = nil
				trigger <- time.Now()
				time.Sleep(100 * time.Millisecond)
			},
			wantBuildCount: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := make(chan time.Time)
			w := newFakeWatcher(c)
			b := &fakeBuilder{}
			r := rebuilder.New(w, b, t.TempDir())
			if err := r.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(r.Stop)
			test.actions(c, *b)
			if b.buildCount != test.wantBuildCount {
				t.Errorf("expected build count %d, got %d", test.wantBuildCount, b.buildCount)
			}

		})
	}
}
