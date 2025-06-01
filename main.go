package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/fivethirty/satisficer/internal/commands"
)

func main() {
	w := os.Stderr
	if err := commands.Execute(w, os.Args, commands.Commands); err != nil {
		_, _ = fmt.Fprintf(w, "\nError: %s\n", err.Error())
		if errors.Is(err, commands.ErrBadCommand) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
