package executor

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const header string = "\n//===[ S A T I S F I C E R ]===\\\\\n\n"

var commands = map[string]*command{
	"create": &create,
	"build":  &build,
}

func Execute(args []string) error {
	mainFlagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	mainFlagSet.Usage = help
	if err := mainFlagSet.Parse(args[1:]); err != nil {
		return nil
	}

	subName := mainFlagSet.Arg(0)
	subFlagSet := flag.NewFlagSet(subName, flag.ContinueOnError)

	subCommand, ok := commands[subName]
	if !ok {
		help()
		return nil
	}

	subFlagSet.Usage = subCommand.usage
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

func help() {
	sb := strings.Builder{}
	sb.WriteString("Usage: satisficer <command> [options]\n\n")
	sb.WriteString("Commands:\n\n")
	sb.WriteString("  create   Create a new project.\n")
	sb.WriteString("  serve    Start a local dev server.\n")
	sb.WriteString("  build    Build a site.\n\n")
	sb.WriteString("Use 'satisficer <command> -h' for more information on a command.\n")
	fmt.Fprint(os.Stderr, sb.String())
}
