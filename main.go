package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/fivethirty/satisficer/internal/commands"
)

func main() {
	if err := commands.Execute(os.Args); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		if errors.Is(err, commands.ErrBadCommand) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
