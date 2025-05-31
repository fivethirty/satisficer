package main

import (
	"fmt"
	"os"

	"github.com/fivethirty/satisficer/internal/commands"
)

func main() {
	if err := commands.Execute(os.Args); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
