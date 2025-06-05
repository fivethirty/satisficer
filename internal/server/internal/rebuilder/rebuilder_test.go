package rebuilder_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/server/internal/rebuilder"
)

const (
	maxSuccessBuildCount = 2
	buildCountFile       = "build_count.txt"
)

type fakeBuilder struct {
	buildCount int
}

func (b *fakeBuilder) Build(buildDir string) error {
	b.buildCount++
	if b.buildCount > maxSuccessBuildCount {
		return errors.New("error building")
	}
	dest := filepath.Join(buildDir, buildCountFile)
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

func (w *fakeWatcher) Ch() <-chan time.Time {
	return w.c
}

func TestRebuilder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		builder          fakeBuilder
		watcherEmitCount int
	}{
		{
			name:             "initial build happens on start",
			builder:          fakeBuilder{},
			watcherEmitCount: 0,
		},
		{
			name:             "watcher triggers rebuild",
			builder:          fakeBuilder{},
			watcherEmitCount: 1,
		},
		{
			name:             "error building does not rotate fs build",
			builder:          fakeBuilder{},
			watcherEmitCount: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := make(chan time.Time)
			w := newFakeWatcher(c)
			b := &fakeBuilder{}
			r, err := rebuilder.Start(t.Context(), w, b, t.TempDir())
			if err != nil {
				t.Fatal(err)
			}

			// extra 1 is for the build during the call to start
			for range test.watcherEmitCount + 1 {
				select {
				case err := <-r.Ch():
					if err == nil {
						if b.buildCount > maxSuccessBuildCount {
							t.Fatalf("expected error on build %d, got: %v", b.buildCount, err)
						}
						expectBuildCount(t, r, b.buildCount)
					} else {
						if b.buildCount <= maxSuccessBuildCount {
							t.Fatalf("expected error on build %d: %v", b.buildCount, err)
						}
						expectBuildCount(t, r, maxSuccessBuildCount)
					}

					if b.buildCount <= test.watcherEmitCount {
						c <- time.Now()
					}
				case <-time.After(100 * time.Millisecond):
					t.Fatalf("timeout waiting for build %d", b.buildCount)
				}
			}
		})
	}
}

func expectBuildCount(t *testing.T, r *rebuilder.Rebuilder, expectedCount int) {
	t.Helper()
	fsys := r.Latest()
	content, err := fs.ReadFile(fsys, buildCountFile)
	if err != nil {
		t.Fatal(err)
	}
	count, err := strconv.Atoi(string(content))
	if err != nil {
		t.Fatal(err)
	}
	if count != expectedCount {
		t.Fatalf("expected build count %d, got %d", expectedCount, count)
	}
}

func TestRebuilderCleansTmpDirs(t *testing.T) {
	t.Parallel()

	c := make(chan time.Time)
	w := newFakeWatcher(c)
	b := &fakeBuilder{}
	baseDir := t.TempDir()
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	r, err := rebuilder.Start(ctx, w, b, baseDir)
	if err != nil {
		t.Fatal(err)
	}

	for i := range 5 {
		select {
		case <-r.Ch():
			if i < 4 {
				fmt.Println(r.BuildDir)
				c <- time.Now()
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout waiting for build %d", i)
		}
	}

	tmpDirs, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Base(r.BuildDir)
	for _, dir := range tmpDirs {
		if dir.IsDir() && dir.Name() != filepath.Base(r.BuildDir) {
			t.Fatalf("expected only active build dir %s, found: %s", expected, dir.Name())
		}
	}

	cancel()

	time.Sleep(100 * time.Millisecond)

	file, err := os.Open(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = file.Close()
	})

	if _, err = file.Readdirnames(1); err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("expected no files in baseDir after cancel, found: %v", err)
	}
}
