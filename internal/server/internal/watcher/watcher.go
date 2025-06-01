package watcher

import (
	"fmt"
	"io/fs"
	"reflect"
	"time"
)

type Watcher struct {
	FSys     fs.FS
	C        chan time.Time
	trigger  chan time.Time
	done     chan bool
	previous map[string]time.Time
	current  map[string]time.Time
}

func New(fsys fs.FS, trigger chan time.Time) (*Watcher, error) {
	w := Watcher{
		FSys:     fsys,
		C:        make(chan time.Time),
		trigger:  trigger,
		done:     make(chan bool),
		previous: make(map[string]time.Time),
		current:  make(map[string]time.Time),
	}
	if err := w.update(); err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}
	return &w, nil
}

func (w *Watcher) Start() error {
	go func() {
		for {
			select {
			case <-w.done:
				return
			case t := <-w.trigger:
				changed, err := w.isChanged()
				if err != nil {
					return
				}
				if !changed {
					continue
				}
				w.C <- t
			}
		}
	}()
	return nil
}

func (w *Watcher) Close() {
	w.done <- true
	close(w.done)
	close(w.C)
}

func (w *Watcher) update() error {
	w.previous = w.current
	w.current = make(map[string]time.Time)
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
		w.current[path] = info.ModTime()
		return nil
	})
}

func (w *Watcher) isChanged() (bool, error) {
	if err := w.update(); err != nil {
		return false, err
	}
	return !reflect.DeepEqual(w.previous, w.current), nil
}
