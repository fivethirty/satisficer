package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/fivethirty/satisficer/internal/commands"
	"github.com/fivethirty/satisficer/internal/logs"
)

func main() {
	w := commands.DefaultWriter
	slog.SetDefault(slog.New(logs.NewHandler(w)))
	if err := commands.Execute(os.Args); err != nil {
		if errors.Is(err, commands.ErrHelp) {
			os.Exit(0)
		}
		_, _ = fmt.Fprintf(w, "\nError: %s\n", err.Error())
		if errors.Is(err, commands.ErrBadCommand) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
