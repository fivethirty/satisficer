package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/server/internal/rebuilder"
)

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
			h, err := rebuilder.New(w, b, t.TempDir())
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
