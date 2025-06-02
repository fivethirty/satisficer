package watcher

import (
	"fmt"
	"io/fs"
	"reflect"
	"sync"
	"time"
)

type state int

const (
	created state = iota
	running
	stopped
)

type Watcher struct {
	FSys       fs.FS
	c          chan time.Time
	trigger    <-chan time.Time
	done       chan bool
	previous   map[string]time.Time
	current    map[string]time.Time
	state      state
	stateMutex sync.Mutex
}

func New(fsys fs.FS, trigger <-chan time.Time) (*Watcher, error) {
	w := Watcher{
		FSys:     fsys,
		trigger:  trigger,
		previous: make(map[string]time.Time),
		current:  make(map[string]time.Time),
		state:    created,
	}
	if err := w.update(); err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}
	return &w, nil
}

func (w *Watcher) C() <-chan time.Time {
	return w.c
}

func (w *Watcher) Start() error {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	if w.state != created {
		return fmt.Errorf("watcher has already been started")
	}
	w.state = running

	w.c = make(chan time.Time)
	w.done = make(chan bool)

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
				w.c <- t
			}
		}
	}()
	return nil
}

func (w *Watcher) Stop() {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	if w.state != running {
		return
	}
	w.done <- true
	close(w.done)
	close(w.c)
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
