package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fivethirty/satisficer/internal/creator"
)

func main() {
	flag.Parse()

	cmd := flag.Arg(0)

	if cmd == "help" || flag.NArg() == 0 {
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

	unknown(cmd)
}

func help() {
	sb := strings.Builder{}
	sb.WriteString("Usage: satisficer <command>\n\n")
	sb.WriteString("Commands:\n")
	sb.WriteString("  create   Create a new site\n")
	sb.WriteString("  serve    Start a local server to serve the site\n")
	sb.WriteString("  build    Generate the static site in the current directory\n")
	sb.WriteString("  help     Show this help message\n\n")
	stdout(sb.String())
}

func unknown(cmd string) {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Unknown command: %s\n", cmd))
	sb.WriteString("Run 'satisficer help' for usage.\n")
	stdout(sb.String())
}

func stdout(s string) {
	if _, err := fmt.Fprint(os.Stdout, s); err != nil {
		exit(err)
	}
}

func exit(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
