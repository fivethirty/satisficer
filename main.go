package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fivethirty/satisficer/internal/executor"
)

func main() {
	if flag.CommandLine == nil {
		os.Exit(1)
	}
	if err := executor.Execute(os.Args); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
