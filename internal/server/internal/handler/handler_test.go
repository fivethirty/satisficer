package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/server/internal/handler"
)

const (
	buildFile = "output.txt"
)

type fakeBuilder struct {
	content string
	err     error
}

func (b *fakeBuilder) Build(buildDir string) error {
	if b.err != nil {
		return b.err
	}
	dest := filepath.Join(buildDir, buildFile)
	return os.WriteFile(dest, []byte(b.content), os.ModePerm)
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
		name  string
		steps []fakeBuilder
	}{
		{
			name: "simple build",
			steps: []fakeBuilder{
				{
					content: "first build",
					err:     nil,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := make(chan time.Time)
			w := newFakeWatcher(c)
			baseDir := t.TempDir()

			h, err := handler.Start(t.Context(), w, &fakeBuilder{}, baseDir)
			if err != nil {
				t.Fatal(err)
			}

			server := httptest.NewServer(h)
			t.Cleanup(server.Close)

			fmt.Printf("Test server listening at %s\n", server.URL)

			// xxx this is blocking the server from closing need to kill it when the ctx times out

			go func() {
				resp, err := http.Get(server.URL + "/_/events")
				if err != nil {
					// do something better
					return
				}
				if resp.StatusCode != http.StatusOK {
					return
				}
			}()

			fmt.Println("Waiting for initial build...")

			time.Sleep(100 * time.Millisecond) // wait for initial build
		})
	}

}

/*func TestRebuilderCleansTmpDirs(t *testing.T) {
	t.Parallel()

	c := make(chan time.Time)
	w := newFakeWatcher(c)
	b := &fakeBuilder{}
	baseDir := t.TempDir()
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	r, err := handler.Start(ctx, w, b, baseDir)
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
}*/
