package executor

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/fivethirty/satisficer/internal/builder"
	"github.com/fivethirty/satisficer/internal/creator"
)

type command struct {
	usage   func()
	numArgs int
	execute func(args []string) error
}

var create = command{
	usage: func() {
		fmt.Fprint(os.Stderr, "Usage: satisficer create <project-dir>\n\n")
		fmt.Fprintln(
			os.Stderr,
			"Creates a new Satisficer project in the specified directory. The directory",
		)
		fmt.Fprintln(os.Stderr, "must not already exist.")
	},
	numArgs: 1,
	execute: func(args []string) error {
		return creator.Create(args[0])
	},
}

var build = command{
	usage: func() {
		fmt.Fprint(os.Stderr, "Usage: satisficer build <project-dir> <output-dir>\n\n")
		fmt.Fprintln(
			os.Stderr,
			"Builds the Satisficer project located in the specified directory and outputs",
		)
		fmt.Fprintln(os.Stderr, "the generated files to the specified output directory.")
	},
	numArgs: 2,
	execute: func(args []string) error {
		projectFS := os.DirFS(args[0])
		layoutFS, err := fs.Sub(projectFS, "layout")
		if err != nil {
			return err
		}
		contentFS, err := fs.Sub(projectFS, "content")
		if err != nil {
			return err
		}

		outputDir := args[1]
		b, err := builder.New(layoutFS, contentFS, outputDir)
		if err != nil {
			return err
		}

		return b.Build()
	},
}
