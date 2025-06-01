package server

import (
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/fivethirty/satisficer/internal/builder"
	"github.com/fivethirty/satisficer/internal/server/internal/watcher"
)

type Server struct {
	Port      int
	ProjectFS fs.FS
	builder   Builder
	watcher   Watcher
}

type Builder interface {
	Build() error
	BuildDir() string
	ProjectFS() fs.FS
}

type Watcher interface {
	C() chan time.Time
	Start() error
	Stop()
}

func New(projectFS fs.FS, port int) (*Server, error) {
	// xxx this is the wrong dir
	builder, err := builder.New(projectFS, "build")
	if err != nil {
		return nil, err
	}
	ticker := time.NewTicker(time.Second)
	watcher, err := watcher.New(projectFS, ticker.C)
	if err != nil {
		return nil, err
	}
	return &Server{
		Port:      port,
		ProjectFS: projectFS,
		builder:   builder,
		watcher:   watcher,
	}, nil
}

func (s *Server) Start() error {
	err := s.builder.Build()
	if err != nil {
		// xxx don't actually fail here, update the state
		return err
	}
	if err != nil {
		return err
	}

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server error:", err)
	}
	return nil
}
