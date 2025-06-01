package server

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/fivethirty/satisficer/internal/builder"
	"github.com/fivethirty/satisficer/internal/server/internal/watcher"
)

func Start(projectFS fs.FS, port int) error {
	outputDir := "some temp dir"
	b, err := builder.New(projectFS, outputDir)
	if err != nil {
		return err
	}
	if err := b.Build(); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)
	watcher, err := watcher.New(projectFS, ticker.C)
	if err != nil {
		return err
	}

	if err := watcher.Start(); err != nil {
		return err
	}
	defer watcher.Stop()

	// xxx: idea: rotate tmp dirs and delete the old ones?
	// this will keep
	go func() {
		for t := range watcher.C {
			log.Println("Change detected at", t)
			if err := b.Build(); err != nil {
				log.Println("Error during build:", err)
			} else {
				log.Println("Build completed successfully")
			}
		}
	}()

	// servce the content using http.FileServer

	portStr := fmt.Sprintf(":%d", port)
	err = http.ListenAndServe(portStr, nil)
	if err != nil {
		log.Fatal("Server error:", err)
	}
	return nil
}
