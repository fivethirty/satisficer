package handler_test

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/server/internal/handler"
)

const (
	dirPerm   = 0o750
	filePerm  = 0o644
	buildFile = "foo/index.html"
)

//go:embed responsebody/html/reload.html
var reloadHTML string

type fakeBuilder struct {
	content string
	err     error
}

func (b *fakeBuilder) Build(buildDir string) error {
	if b.err != nil {
		return b.err
	}
	dest := filepath.Join(buildDir, buildFile)
	if err := os.MkdirAll(filepath.Dir(dest), dirPerm); err != nil {
		return err
	}
	return os.WriteFile(dest, []byte(b.content), filePerm)
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

func TestHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		steps []fakeBuilder
	}{
		{
			name:  "initial build only",
			steps: []fakeBuilder{},
		},
		{
			name: "many builds",
			steps: []fakeBuilder{
				{content: "build 1"},
			},
		},
		{
			name: "build with error",
			steps: []fakeBuilder{
				{err: fmt.Errorf("build error")},
			},
		},
		{
			name: "recovery after error",
			steps: []fakeBuilder{
				{err: fmt.Errorf("build error")},
				{content: "build 2"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			watcherCh := make(chan time.Time)
			w := newFakeWatcher(watcherCh)
			baseDir := t.TempDir()
			fb := &fakeBuilder{
				content: "initial build",
			}

			h, err := handler.Start(t.Context(), w, fb, baseDir)
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(h)
			t.Cleanup(server.Close)
			testRequest(t, httptest.NewServer(h), fb)

			sCh := sseCh(t, server)

			for _, step := range test.steps {
				*fb = step
				watcherCh <- time.Now()
				select {
				case err := <-sCh:
					if err != nil {
						t.Fatalf("error from SSE client: %v", err)
					}
					testRequest(t, server, fb)
				case <-time.After(100 * time.Millisecond):
					t.Fatalf("timeout waiting for rebuild event")
				}
			}
		})
	}
}

func TestHandler_TrailingSlash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "no trailing slash",
			url:  "/foo",
		},
		{
			name: "with trailing slash",
			url:  "/foo/",
		},
		{
			name: "with double trailing slash",
			url:  "/foo//",
		},
		{
			name: "full path",
			url:  "/foo/index.html",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			watcherCh := make(chan time.Time)
			w := newFakeWatcher(watcherCh)
			baseDir := t.TempDir()
			fb := &fakeBuilder{
				content: "initial build",
			}

			h, err := handler.Start(t.Context(), w, fb, baseDir)
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(h)
			t.Cleanup(server.Close)

			resp, err := http.Get(server.URL + test.url)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got %v", resp.StatusCode)
			}
		})
	}
}

func testRequest(t *testing.T, server *httptest.Server, fb *fakeBuilder) {
	t.Helper()
	resp, err := http.Get(server.URL + "/" + buildFile)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if fb.err != nil {
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status InternalServerError, got %v", resp.StatusCode)
		}
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body: %v", err)
	}
	expected := fmt.Sprintf("%s%s", reloadHTML, fb.content)
	bodyStr := string(body)
	if bodyStr != expected {
		t.Fatalf("expected build content '%s', got '%s'", expected, bodyStr)
	}
}

func sseCh(t *testing.T, server *httptest.Server) <-chan error {
	t.Helper()
	ch := make(chan error, 1)
	go func() {
		resp, err := http.Get(server.URL + "/_/events")
		if err != nil {
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-t.Context().Done():
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					if err.Error() == "EOF" {
						break
					} else {
						ch <- fmt.Errorf("error reading event stream: %v", err)
						return
					}
				}
				if line == "\n" {
					continue
				}
				if line != "data: rebuild\n" {
					ch <- fmt.Errorf("unexpected event format: %s", line)
					return
				}
				ch <- nil
			}
		}
	}()
	return ch
}

func TestHandler_RemovesTempFilesOnRebuild(t *testing.T) {
	t.Parallel()

	watcherCh := make(chan time.Time)
	w := newFakeWatcher(watcherCh)
	baseDir := t.TempDir()
	fb := &fakeBuilder{
		content: "initial build",
	}

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	h, err := handler.Start(ctx, w, fb, baseDir)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(h)
	t.Cleanup(server.Close)

	watcherCh <- time.Now()

	files, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatalf("error reading baseDir: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected one file after build, found %d", len(files))
	}
}
