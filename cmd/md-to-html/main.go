package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/fserg/md-to-html/internal/converter"
	"github.com/fserg/md-to-html/internal/server"
	"github.com/fserg/md-to-html/internal/version"
	webtemplate "github.com/fserg/md-to-html/web/template"
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
		return runServe(args[1:], stdout, stderr)
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

func runServe(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: md-to-html serve")
		return 2
	}

	cfg, err := server.LoadConfig()
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}

	conv, err := converter.New(webtemplate.FS)
	if err != nil {
		fmt.Fprintf(stderr, "load converter: %v\n", err)
		return 1
	}

	srv, err := server.New(cfg, conv)
	if err != nil {
		fmt.Fprintf(stderr, "create server: %v\n", err)
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := srv.Run(ctx); err != nil {
		fmt.Fprintf(stderr, "run server: %v\n", err)
		return 1
	}

	_ = stdout
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
  serve    Start the HTTP server
  cli      Convert a Markdown file stub
  version  Print the build version
`)
}
