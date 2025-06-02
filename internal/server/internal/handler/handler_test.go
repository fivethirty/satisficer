package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		name       string
		path       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "can get built file",
			path:       "/build_count.txt",
			err:        nil,
			wantStatus: 200,
			wantBody:   "1",
		},
		{
			name:       "error building site",
			path:       "/build_count.txt",
			err:        errors.New("bad build"),
			wantStatus: 500,
			wantBody:   "Error building site, please see terminal logs for more info: error building site: bad build",
		},
		{
			name:       "file not found",
			path:       "/non_existent_file.txt",
			err:        nil,
			wantStatus: 404,
			wantBody:   "404 page not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := make(chan time.Time)
			w := newFakeWatcher(c)
			b := &fakeBuilder{
				err: test.err,
			}
			h, err := handler.New(w, b, t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(h.Stop)
			req, err := http.NewRequest("GET", test.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != test.wantStatus {
				t.Errorf("expected status %d, got %d", test.wantStatus, rec.Code)
			}
			actual := strings.TrimSpace(rec.Body.String())
			if actual != test.wantBody {
				t.Errorf("expected body %q, got %q", test.wantBody, actual)
			}
		})
	}
}

func TestHandlerRebuild(t *testing.T) {
	t.Parallel()

	c := make(chan time.Time)
	w := newFakeWatcher(c)
	b := &fakeBuilder{}
	h, err := handler.New(w, b, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.Stop)

	req, err := http.NewRequest("GET", "/build_count.txt", nil)
	if err != nil {
		t.Fatal(err)
	}

	testGoodBuild(t, h, req, 1)

	c <- time.Now()
	time.Sleep(100 * time.Millisecond)
	testGoodBuild(t, h, req, 2)

	b.err = errors.New("bad build")
	c <- time.Now()
	time.Sleep(100 * time.Millisecond)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	b.err = nil
	c <- time.Now()
	time.Sleep(100 * time.Millisecond)
	testGoodBuild(t, h, req, 3)
}

func testGoodBuild(t *testing.T, h *handler.Handler, req *http.Request, expectedCount int) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != strconv.Itoa(expectedCount) {
		t.Fatalf("expected body '%d', got %s", expectedCount, rec.Body.String())
	}
}
