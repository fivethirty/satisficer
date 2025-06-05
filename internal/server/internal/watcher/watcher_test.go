package watcher_test

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/fivethirty/satisficer/internal/server/internal/watcher"
)

func TestWatcher(t *testing.T) {
	t.Parallel()

	t0 := time.Now()
	t1 := t0.Add(1 * time.Second)

	tests := []struct {
		name        string
		t1State     fstest.MapFS
		t2State     fstest.MapFS
		expectEvent bool
	}{
		{
			name: "no changes",
			t1State: fstest.MapFS{
				"file1.txt": &fstest.MapFile{
					ModTime: t0,
				},
			},
			t2State: fstest.MapFS{
				"file1.txt": &fstest.MapFile{
					ModTime: t0,
				},
			},
			expectEvent: false,
		},
		{
			name: "file modified",
			t1State: fstest.MapFS{
				"file1.txt": &fstest.MapFile{
					ModTime: t0,
				},
			},
			t2State: fstest.MapFS{
				"file1.txt": &fstest.MapFile{
					ModTime: t1,
				},
			},
			expectEvent: true,
		},
		{
			name: "file added",
			t1State: fstest.MapFS{
				"file1.txt": &fstest.MapFile{
					ModTime: t0,
				},
			},
			t2State: fstest.MapFS{
				"file1.txt": &fstest.MapFile{
					ModTime: t0,
				},
				"sub/file2.txt": &fstest.MapFile{
					ModTime: t1,
				},
			},
			expectEvent: true,
		},
		{
			name: "file deleted",
			t1State: fstest.MapFS{
				"file1.txt": &fstest.MapFile{
					ModTime: t0,
				},
			},
			t2State:     fstest.MapFS{},
			expectEvent: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			channel := make(chan time.Time)
			t.Cleanup(func() {
				close(channel)
			})

			w, err := watcher.Start(t.Context(), test.t1State, channel)
			if err != nil {
				t.Fatal(err)
			}

			w.FSys = test.t2State

			channel <- t1

			select {
			case eventTime := <-w.Ch():
				if !test.expectEvent {
					t.Fatal("expected no event, but got one")
				}
				if eventTime != t1 {
					t.Fatalf("expected event time %v, got %v", t1, eventTime)
				}
			case <-time.After(100 * time.Millisecond):
				if test.expectEvent {
					t.Fatal("expected an event, but got none")
				}
			}
		})
	}

}
