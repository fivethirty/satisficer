package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fivethirty/satisficer/internal/commands"
)

func main() {
	if flag.CommandLine == nil {
		os.Exit(1)
	}
	if err := commands.Run(os.Args); err != nil {
		if err == commands.ErrInvalidUsage {
			flag.CommandLine.Usage()
			os.Exit(0)
		}
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

/*func main() {
	flag.Parse()

	cmd := flag.Arg(0)

	if cmd == "help" || flag.NArg() == 0 {
		stdout("\n//===[ S A T I S F I C E R ]===\\\\\n\n")
		help()
		os.Exit(0)
	}

	if cmd == "create" {
		if flag.NArg() != 2 {
			stdout("Usage: satisficer new <directory>\n")
			os.Exit(0)
		}
		stdout("\n//===[ S A T I S F I C E R ]===\\\\\n\n")
		dir := flag.Arg(1)
		err := creator.Create(dir)
		if err != nil {
			exit(err)
		}
		os.Exit(0)
	}

	if cmd == "build" {
		if flag.NArg() != 3 {
			stdout("Usage: satisficer build <project-dir> <output-dir>\n")
			os.Exit(0)
		}
		stdout("\n//===[ S A T I S F I C E R ]===\\\\\n\n")
		os.Exit(0)
	}

	unknown(cmd)
}

func help() {
	sb := strings.Builder{}
	sb.WriteString("Usage: satisficer <command>\n\n")
	sb.WriteString("Commands:\n")
	sb.WriteString("  create   Create a new site\n")
	sb.WriteString("  serve    Start a local dev server\n")
	sb.WriteString("  build    Build a site\n")
	sb.WriteString("  help     Show this help message\n\n")
	stdout(sb.String())
}

func unknown(cmd string) {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Unknown command: %s\n", cmd))
	sb.WriteString("Run 'satisficer help' for usage.\n")
	stdout(sb.String())
}*/
