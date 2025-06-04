package handler

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type FileServerProvider interface {
	FileServer() http.Handler
	// xxx rename all these to something useful
	C() <-chan time.Time
}

type Handler stuct {
	atomicHandler  atomic.Value
}
func (r *Rebuilder) FileServer() http.Handler {
	if v := r.atomicHandler.Load(); v != nil {
		return v.(http.Handler)
	}
	return nil
}

func (r *Rebuilder) rotateHandler() error {
	if err := r.buildAndCleanup(); err != nil {
		return err
	}
	r.atomicHandler.Store(http.FileServer(http.Dir(r.activeBuildDir)))
	return nil
}

func handler() (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /_/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		for {
			// consume changed events and spit them out as server-sent events
			return
		}
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if h.currentError != nil {
			msg := fmt.Sprintf(
				"Error building site, please see terminal logs for more info: %v",
				h.currentError,
			)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		successHandler := h.atomicSuccessHandler.Load()
		if successHandler == nil {
			http.Error(w, "Site is not ready yet", http.StatusServiceUnavailable)
			return
		}

		successHandler.(http.Handler).ServeHTTP(w, r)
	})
	return mux, nil
}
