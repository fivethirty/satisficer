package watcher

import (
	"context"
	"fmt"
	"io/fs"
	"reflect"
	"time"
)

type Watcher struct {
	FSys          fs.FS
	changedCh     chan time.Time
	previousFiles map[string]time.Time
	currentFiles  map[string]time.Time
}

func Start(ctx context.Context, fsys fs.FS, triggerCh <-chan time.Time) (*Watcher, error) {
	w := Watcher{
		FSys:          fsys,
		changedCh:     make(chan time.Time, 1),
		previousFiles: make(map[string]time.Time),
		currentFiles:  make(map[string]time.Time),
	}
	if err := w.update(); err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(w.changedCh)
				return
			case t := <-triggerCh:
				isChanged, err := w.isChanged()
				if err != nil {
					return
				}
				if !isChanged {
					continue
				}
				w.publishChanged(t)
			}
		}
	}()

	return &w, nil
}

func (w *Watcher) ChangedCh() <-chan time.Time {
	return w.changedCh
}

func (w *Watcher) publishChanged(t time.Time) {
	select {
	case w.changedCh <- t:
	default:
	}
}

func (w *Watcher) update() error {
	w.previousFiles = w.currentFiles
	w.currentFiles = make(map[string]time.Time)
	return fs.WalkDir(w.FSys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		w.currentFiles[path] = info.ModTime()
		return nil
	})
}

func (w *Watcher) isChanged() (bool, error) {
	if err := w.update(); err != nil {
		return false, err
	}
	return !reflect.DeepEqual(w.previousFiles, w.currentFiles), nil
}
