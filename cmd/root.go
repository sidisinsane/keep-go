// Package cmd implements the keep CLI subcommands.
package cmd

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Run dispatches CLI arguments to the appropriate subcommand.
func Run(args []string) {
	fs := flag.NewFlagSet("keep", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: keep [options] <command>\n\nCommands:\n  graph   Rebuild .keep/graph.json\n  index   Rebuild index.md\n  lint    Lint workspace documents\n  state   Update .keep/state.json\n\nOptions:\n")
		fs.PrintDefaults()
	}
	verbose := fs.Bool("v", false, "Enable debug logging.")
	verboseLong := fs.Bool("verbose", false, "Enable debug logging.")
	_ = fs.Parse(args[1:])

	if *verbose || *verboseLong {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	ws, err := resolveWorkspace()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch remaining[0] {
	case "graph":
		if err := runGraph(ws); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "index":
		if err := runIndex(ws); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "lint":
		if err := runLint(ws); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "state":
		if err := runState(ws); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "error: unknown subcommand %q\n", remaining[0])
		fs.Usage()
		os.Exit(1)
	}
}

func resolveWorkspace() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for _, name := range []string{"keep.yml", "keep.yaml", "keep.json"} {
		if _, err := os.Stat(filepath.Join(cwd, name)); err == nil {
			return cwd, nil
		}
	}
	return "", fmt.Errorf("no keep config file found in %s (keep.yml, keep.yaml, or keep.json)\nMake sure you are running 'keep' from your workspace root.", cwd)
}
