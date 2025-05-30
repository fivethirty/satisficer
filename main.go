package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	flag.Parse()

	out("\n//===[ S A T I S F I C E R ]===\\\\\n\n")

	cmd := flag.Arg(0)

	if cmd == "help" || flag.NArg() == 0 {
		help()
		os.Exit(0)
	}

	if cmd == "new" {
		if flag.NArg() != 2 {
			out("Usage: satisficer new <directory>\n")
			os.Exit(0)
		}
		dir := flag.Arg(1)
		new(dir)
	}

	unknown(cmd)
}

func help() {
	sb := strings.Builder{}
	sb.WriteString("Usage: satisficer <command>\n\n")
	sb.WriteString("Commands:\n")
	sb.WriteString("  new      Create a new site\n")
	sb.WriteString("  serve    Start a local server to serve the site\n")
	sb.WriteString("  build    Generate the static site in the current directory\n")
	sb.WriteString("  help     Show this help message\n\n")
	out(sb.String())
}

//go:embed default
var site embed.FS

func new(dir string) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		error(fmt.Sprintf("Error creating directory: %v\n", err))
	}
}

func unknown(cmd string) {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Unknown command: %s\n", cmd))
	sb.WriteString("Run 'satisficer help' for usage.\n")
	out(sb.String())
}

func out(s string) {
	_, err := os.Stdout.WriteString(s)
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}

func error(s string) {
	_, _ = os.Stderr.WriteString(s)
	os.Exit(1)
}
