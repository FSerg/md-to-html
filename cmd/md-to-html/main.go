package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/fserg/md-to-html/internal/version"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0
	case "serve":
		return runServe(args[1:], stdout)
	case "cli":
		return runCLI(args[1:], stdout, stderr)
	case "version":
		return runVersion(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown subcommand %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runServe(args []string, stdout io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	fmt.Fprintln(stdout, "serve not implemented yet")
	return 0
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("cli", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: md-to-html cli <file.md>")
		return 2
	}

	fmt.Fprintf(stdout, "cli not implemented yet: %s\n", fs.Arg(0))
	return 0
}

func runVersion(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: md-to-html version")
		return 2
	}

	fmt.Fprintln(stdout, version.Version)
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  md-to-html serve
  md-to-html cli <file.md>
  md-to-html version

Commands:
  serve    Start the HTTP server stub
  cli      Convert a Markdown file stub
  version  Print the build version
`)
}
