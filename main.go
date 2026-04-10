package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/example/switcheroo/internal/app"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return app.RunInteractive()
	}

	switch args[0] {
	case "import":
		if len(args) < 2 {
			return errors.New("usage: switcheroo import <git-url>")
		}
		return app.ImportRepo(args[1])
	case "list":
		return app.ListProfiles()
	case "use":
		fs := flag.NewFlagSet("use", flag.ContinueOnError)
		scope := fs.String("scope", "", "apply scope: local or global")
		force := fs.Bool("force", false, "overwrite conflicting files")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		rest := fs.Args()
		if len(rest) < 1 {
			return errors.New("usage: switcheroo use [--scope local|global] [--force] <profile-id>")
		}
		if *scope != "" {
			s := strings.ToLower(*scope)
			if s != "local" && s != "global" {
				return errors.New("--scope must be local or global")
			}
		}
		return app.UseProfile(rest[0], *scope, *force)
	case "help", "-h", "--help":
		printHelp()
		return nil
	default:
		printHelp()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printHelp() {
	fmt.Println(`switcheroo - switch Claude Code/Codex setups

Commands:
  import <git-url>                           Import profiles from a git repository
  list                                       List imported profiles
  use [--scope local|global] [--force] <id> Apply a profile

Run without arguments for interactive mode.`)
}
