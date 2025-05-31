package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/fivethirty/satisficer/internal/commands"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	if err := commands.Execute(os.Args); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
