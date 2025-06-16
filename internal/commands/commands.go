package commands

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/fivethirty/satisficer/internal/builder"
	"github.com/fivethirty/satisficer/internal/creator"
	"github.com/fivethirty/satisficer/internal/server"
)

var DefaultWriter io.Writer = os.Stderr

type Command struct {
	Name      string
	UsageText string
	FlagSet   *flag.FlagSet
	Validate  func() error
	Run       func() error
}

// Bit of a hack that these are empty, but they allow us to figure out
// when to exit with a specific error code upstream without polluting
// the log message.
var ErrBadCommand = errors.New("")
var ErrHelp = errors.New("")

func (sc *Command) usage(err error) error {
	_, _ = fmt.Fprintln(DefaultWriter, sc.UsageText)
	if err != nil {
		return fmt.Errorf("%w%w", err, ErrBadCommand)
	}
	return ErrHelp
}

func (sc *Command) parse(args []string) error {
	if err := sc.FlagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	return nil
}

func (sc *Command) Execute(args []string) error {
	if err := sc.parse(args); err != nil {
		return sc.usage(err)
	}
	if err := sc.Validate(); err != nil {
		return err
	}
	return sc.Run()
}

func (sc *Command) verifyArgCount(expected int) error {
	numArgs := sc.FlagSet.NArg()
	if numArgs != expected {
		if numArgs == 0 {
			return sc.usage(nil)
		}
		return sc.usage(
			fmt.Errorf(
				"expected %d arguments, got %d",
				expected,
				sc.FlagSet.NArg(),
			),
		)
	}
	return nil
}

//go:embed usage
var usageFS embed.FS

func readUsageText(path string) string {
	usage, err := usageFS.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(usage))
}

func flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

var SubCommands = map[string]*Command{
	"create": func() *Command {
		fs := flagSet("create")
		c := &Command{
			UsageText: readUsageText("usage/create.txt"),
			FlagSet:   fs,
		}
		c.Validate = func() error {
			return c.verifyArgCount(1)
		}
		c.Run = func() error {
			return creator.Create(fs.Arg(0))
		}
		return c
	}(),
	"build": func() *Command {
		fs := flagSet("build")
		c := &Command{
			UsageText: readUsageText("usage/build.txt"),
			FlagSet:   fs,
		}
		c.Validate = func() error {
			return c.verifyArgCount(2)
		}
		c.Run = func() error {
			projectFS := os.DirFS(fs.Arg(0))
			buildDir := fs.Arg(1)
			b, err := builder.New(projectFS)
			if err != nil {
				return err
			}

			return b.Build(buildDir)
		}
		return c
	}(),
	"serve": func() *Command {
		fs := flagSet("serve")
		var port uint
		fs.UintVar(&port, "port", 3000, "")
		fs.UintVar(&port, "p", 3000, "")
		c := &Command{
			UsageText: readUsageText("usage/serve.txt"),
			FlagSet:   fs,
		}
		c.Validate = func() error {
			return c.verifyArgCount(1)
		}
		c.Run = func() error {
			projectFS := os.DirFS(fs.Arg(0))
			port := uint16(port)
			return server.Serve(projectFS, port)
		}
		return c
	}(),
}

var version = ""

func getVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}

	if info.Main.Version == "" {
		return "(devel)"
	}

	return info.Main.Version
}

func mainCommand(name string) *Command {
	fs := flagSet(name)
	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "")
	fs.BoolVar(&showVersion, "v", false, "")
	c := &Command{
		UsageText: readUsageText("usage/main.txt"),
		FlagSet:   fs,
		Validate: func() error {
			return nil
		},
	}

	c.Run = func() error {
		if showVersion {
			_, _ = fmt.Fprintf(DefaultWriter, "satisficer version %s\n", getVersion())
			return nil
		}
		subName := fs.Arg(0)
		if subName == "" {
			return c.usage(nil)
		}
		sub, ok := SubCommands[subName]
		if !ok {
			return c.usage(fmt.Errorf("unknown command %q", subName))
		}
		return sub.Execute(fs.Args()[1:])
	}

	return c
}

func Execute(args []string) error {
	cmd := mainCommand(args[0])
	return cmd.Execute(args[1:])
}
