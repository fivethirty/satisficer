package commands

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"

	"github.com/fivethirty/satisficer/internal/builder"
	"github.com/fivethirty/satisficer/internal/creator"
)

const header string = "\n//===[ S A T I S F I C E R ]===\\\\\n\n"

type command struct {
	usage   string
	numArgs int
	execute func(args []string) error
}

//go:embed usage
var usageFS embed.FS

var commands = map[string]*command{
	"create": {
		usage:   "usage/create.txt",
		numArgs: 1,
		execute: func(args []string) error {
			return creator.Create(args[0])
		},
	},
	"build": {
		usage:   "usage/build.txt",
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
	},
}

const usagePath = "usage/main.txt"

func Execute(args []string) error {
	mainFlagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	mainUsage, err := usageFunc(usagePath)
	if err != nil {
		return err
	}

	mainFlagSet.Usage = mainUsage
	if err := mainFlagSet.Parse(args[1:]); err != nil {
		return nil
	}

	subName := mainFlagSet.Arg(0)
	subFlagSet := flag.NewFlagSet(subName, flag.ContinueOnError)

	subCommand, ok := commands[subName]
	if !ok {
		mainUsage()
		return nil
	}

	subUsage, err := usageFunc(subCommand.usage)
	if err != nil {
		return err
	}
	subFlagSet.Usage = subUsage

	if err := subFlagSet.Parse(args[2:]); err != nil {
		return nil
	}
	if subFlagSet.NArg() != subCommand.numArgs {
		subFlagSet.Usage()
		return nil
	}
	fmt.Print(header)
	return subCommand.execute(subFlagSet.Args())
}

func usageFunc(path string) (func(), error) {
	usage, err := usageFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return func() {
		fmt.Println(string(usage))
	}, nil
}
