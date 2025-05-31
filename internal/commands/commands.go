package commands

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/fivethirty/satisficer/internal/builder"
	"github.com/fivethirty/satisficer/internal/creator"
)

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

const (
	mainUsagePath = "usage/main.txt"
	header        = "\n//===[ S A T I S F I C E R ]===\\\\\n\n"
)

func Execute(args []string) error {
	mainFlagSet := flagSet(args[0])
	mainUsage, err := usage(mainUsagePath)
	if err != nil {
		return err
	}
	if err := parse(mainFlagSet, args[1:], mainUsage); err != nil {
		return err
	}

	subName := mainFlagSet.Arg(0)
	if subName == "" {
		return mainUsage(nil)
	}
	subFlagSet := flagSet(subName)

	subCommand, ok := commands[subName]
	if !ok {
		return mainUsage(fmt.Errorf("unknown command: %s%w", subName, ErrBadCommand))
	}

	subUsage, err := usage(subCommand.usage)
	if err != nil {
		return err
	}

	if err := parse(subFlagSet, args[2:], subUsage); err != nil {
		return err
	}
	if subFlagSet.NArg() != subCommand.numArgs {
		return subUsage(
			fmt.Errorf(
				"expected %d arguments, got %d%w",
				subCommand.numArgs,
				subFlagSet.NArg(),
				ErrBadCommand,
			),
		)
	}
	fmt.Print(header)
	return subCommand.execute(subFlagSet.Args())
}

func flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

type usageFunc func(error) error

func usage(usagePath string) (usageFunc, error) {
	usage, err := usageFS.ReadFile(usagePath)
	if err != nil {
		return nil, err
	}
	return func(err error) error {
		fmt.Println(string(usage))
		return err
	}, nil
}

// Bit of a hack that this is empty, but it allows us to figure out
// when to exit with a specific error code upstream without polluting
// the log message.
var ErrBadCommand = errors.New("")

func parse(fs *flag.FlagSet, args []string, uf usageFunc) error {
	if err := fs.Parse(args); err != nil {
		return uf(fmt.Errorf("%w%w", ErrBadCommand, err))
	}
	return nil
}
