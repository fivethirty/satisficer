package commands

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/fivethirty/satisficer/internal/builder"
	"github.com/fivethirty/satisficer/internal/creator"
	"github.com/fivethirty/satisficer/internal/logs"
	"github.com/fivethirty/satisficer/internal/server"
)

type Command struct {
	Usage   string
	NumArgs int
	Execute func(args []string) error
}

//go:embed usage
var usageFS embed.FS

var Commands = map[string]*Command{
	"create": {
		Usage:   "usage/create.txt",
		NumArgs: 1,
		Execute: func(args []string) error {
			return creator.Create(args[0])
		},
	},
	"build": {
		Usage:   "usage/build.txt",
		NumArgs: 2,
		Execute: func(args []string) error {
			projectFS := os.DirFS(args[0])
			buildDir := args[1]
			b, err := builder.New(projectFS)
			if err != nil {
				return err
			}

			return b.Build(buildDir)
		},
	},
	"serve": {
		Usage:   "usage/serve.txt",
		NumArgs: 1,
		Execute: func(args []string) error {
			projectFS := os.DirFS(args[0])
			port := uint16(8080) // Default port, can be changed later
			return server.Serve(projectFS, port)
		},
	},
}

const (
	mainUsagePath = "usage/main.txt"
	header        = "\n//===[ S A T I S F I C E R ]===\\\\\n\n"
)

func Execute(w io.Writer, args []string, commands map[string]*Command) error {
	setLogger(w)
	mainFlagSet := flagSet(args[0])
	mainUsage, err := usage(w, mainUsagePath)
	if err != nil {
		return err
	}
	err = parse(mainFlagSet, args[1:], mainUsage)
	if err != nil {
		return ignoreErrHelp(err)
	}

	subName := mainFlagSet.Arg(0)
	if subName == "" {
		return mainUsage(nil)
	}
	subCommand, ok := commands[subName]
	if !ok {
		return mainUsage(fmt.Errorf("unknown command: %s", subName))
	}

	subUsage, err := usage(w, subCommand.Usage)
	if err != nil {
		return err
	}

	subFlagSet := flagSet(subName)

	err = parse(subFlagSet, mainFlagSet.Args()[1:], subUsage)
	if err != nil {
		return ignoreErrHelp(err)
	}
	numArgs := subFlagSet.NArg()
	if numArgs != subCommand.NumArgs {
		if numArgs == 0 {
			return subUsage(nil)
		}
		return subUsage(
			fmt.Errorf(
				"expected %d arguments, got %d",
				subCommand.NumArgs,
				subFlagSet.NArg(),
			),
		)
	}
	_, _ = fmt.Fprint(w, header)
	return subCommand.Execute(subFlagSet.Args())
}

func setLogger(w io.Writer) {
	slog.SetDefault(slog.New(logs.NewHandler(w)))
}

func flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

type usageFunc func(error) error

func usage(w io.Writer, usagePath string) (usageFunc, error) {
	usage, err := usageFS.ReadFile(usagePath)
	if err != nil {
		return nil, err
	}
	usageStr := strings.TrimSpace(string(usage))
	return func(err error) error {
		_, _ = fmt.Fprintln(w, usageStr)
		if err != nil {
			return fmt.Errorf("%w%w", err, ErrBadCommand)
		}
		return nil
	}, nil
}

// Bit of a hack that this is empty, but it allows us to figure out
// when to exit with a specific error code upstream without polluting
// the log message.
var ErrBadCommand = errors.New("")

func parse(fs *flag.FlagSet, args []string, uf usageFunc) error {
	if err := fs.Parse(args); err != nil {
		return uf(err)
	}
	return nil
}

func ignoreErrHelp(err error) error {
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	return err
}
