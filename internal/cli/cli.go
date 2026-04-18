package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fserg/md-to-html/internal/converter"
	webtemplate "github.com/fserg/md-to-html/web/template"
)

var ErrUsage = errors.New("cli usage error")

func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if wantsHelp(args) {
		printUsage(stdout)
		return nil
	}

	normalized, err := normalizeArgs(args)
	if err != nil {
		printUsage(stderr)
		return fmt.Errorf("%w: %v", ErrUsage, err)
	}

	fs := flag.NewFlagSet("cli", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		output   string
		title    string
		useStdin bool
	)

	fs.StringVar(&output, "output", "", "output file path")
	fs.StringVar(&output, "o", "", "output file path")
	fs.StringVar(&title, "title", "", "fallback title if markdown has no headings")
	fs.BoolVar(&useStdin, "stdin", false, "read markdown from stdin")

	if err := fs.Parse(normalized); err != nil {
		printUsage(stderr)
		return fmt.Errorf("%w: %v", ErrUsage, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	positional := fs.Args()
	if len(positional) > 1 {
		printUsage(stderr)
		return fmt.Errorf("%w: expected a single input file or '-'", ErrUsage)
	}

	conv, err := converter.New(webtemplate.FS)
	if err != nil {
		return fmt.Errorf("init converter: %w", err)
	}

	var (
		markdown      []byte
		fallbackTitle = title
		outputPath    = output
		writeToStdout bool
	)

	switch {
	case useStdin || (len(positional) == 1 && positional[0] == "-"):
		markdown, err = io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		if fallbackTitle == "" {
			fallbackTitle = "Document"
		}
		writeToStdout = outputPath == ""
	case len(positional) == 1:
		inputPath := positional[0]
		markdown, err = os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", inputPath, err)
		}
		if fallbackTitle == "" {
			fallbackTitle = strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		}
		if outputPath == "" {
			outputPath = strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".html"
		}
	default:
		printUsage(stderr)
		return fmt.Errorf("%w: no input specified", ErrUsage)
	}

	result, err := conv.Convert(markdown, fallbackTitle)
	if err != nil {
		return fmt.Errorf("convert markdown: %w", err)
	}

	if writeToStdout {
		_, err = stdout.Write(result.HTML)
		if err != nil {
			return fmt.Errorf("write stdout: %w", err)
		}
		return nil
	}

	if err := os.WriteFile(outputPath, result.HTML, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outputPath, err)
	}

	return nil
}

func normalizeArgs(args []string) ([]string, error) {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, 1)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			positionals = append(positionals, args[i+1:]...)
			return append(flags, positionals...), nil
		case arg == "-":
			positionals = append(positionals, arg)
		case !strings.HasPrefix(arg, "-"):
			positionals = append(positionals, arg)
		case strings.HasPrefix(arg, "--output="), strings.HasPrefix(arg, "--title="), strings.HasPrefix(arg, "-o="):
			flags = append(flags, arg)
		case arg == "--output" || arg == "-o" || arg == "--title":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag needs an argument: %s", arg)
			}
			flags = append(flags, arg, args[i+1])
			i++
		default:
			flags = append(flags, arg)
		}
	}

	return append(flags, positionals...), nil
}

func wantsHelp(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "-h", "--help", "-help":
			return true
		}
	}
	return false
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `Usage: md-to-html cli [--stdin|-|<file.md>] [--output path] [--title str]

Options:
  --stdin        Read markdown from stdin
  -o, --output   Output file path (default: stdout for stdin, <input>.html for file)
  --title        Fallback title if markdown has no headings
  -h, --help     Show this help
`)
}
