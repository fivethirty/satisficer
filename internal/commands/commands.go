package commands

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/fivethirty/satisficer/internal/creator"
	"github.com/fivethirty/satisficer/internal/generator"
)

const header string = "\n//===[ S A T I S F I C E R ]===\\\\\n\n"

var ErrInvalidUsage = errors.New("invalid usage")

func Run(args []string) error {
	main := flag.NewFlagSet(args[0], flag.ContinueOnError)
	main.Usage = usage
	if err := main.Parse(args[1:]); err != nil {
		return nil
	}

	subCmd := main.Arg(0)
	sub := flag.NewFlagSet(subCmd, flag.ContinueOnError)

	switch subCmd {
	case "create":
		sub.Usage = func() {
			fmt.Fprint(os.Stderr, "Usage: satisficer create <project-dir>\n\n")
			fmt.Fprintln(
				os.Stderr,
				"Creates a new Satisficer project in the specified directory. The directory",
			)
			fmt.Fprintln(os.Stderr, "must not already exist.")
		}
		if err := sub.Parse(args[2:]); err != nil {
			return nil
		}
		if sub.NArg() != 1 {
			sub.Usage()
			return nil
		}
		fmt.Print(header)
		return creator.Create(sub.Arg(0))
	case "build":
		sub.Usage = func() {
			fmt.Fprint(os.Stderr, "Usage: satisficer build <project-dir> <output-dir>\n\n")
			fmt.Fprintln(
				os.Stderr,
				"Builds the Satisficer project located in the specified directory and outputs",
			)
			fmt.Fprintln(os.Stderr, "the built site to the specified output directory.")
		}
		if err := sub.Parse(args[2:]); err != nil {
			return nil
		}
		if sub.NArg() != 2 {
			sub.Usage()
			return nil
		}
		fmt.Print(header)
		projectFS := os.DirFS(sub.Arg(0))

		layoutFS, err := fs.Sub(projectFS, "layout")
		if err != nil {
			return err
		}
		contentFS, err := fs.Sub(projectFS, "content")
		if err != nil {
			return err
		}

		outputDir := sub.Arg(1)
		g, err := generator.New(layoutFS, contentFS, outputDir)
		if err != nil {
			return err
		}

		return g.Generate()
	}

	main.Usage()
	return nil
}

func usage() {
	sb := strings.Builder{}
	sb.WriteString("Usage: satisficer <command> [options]\n\n")
	sb.WriteString("Commands:\n\n")
	sb.WriteString("  create   Create a new project.\n")
	sb.WriteString("  serve    Start a local dev server.\n")
	sb.WriteString("  build    Build a site.\n\n")
	sb.WriteString("Use 'satisficer <command> -h' for more information on a command.\n")
	fmt.Fprint(os.Stderr, sb.String())
}
