// temgo is a focused work timer for the terminal.
// It supports CLI and TUI modes with JSON-based plans,
// session history, statistics, and desktop notifications.
package main

import (
	"fmt"
	"os"

	"github.com/venexene/temgo/internal/commands"
	"github.com/venexene/temgo/internal/plan"
)

// version is set at build time via ldflags.
var version = "dev"

const mainUsage = `Usage: temgo <command> [arguments]

A focused work timer with JSON plans and session history.

Commands:
  start     Start a timer in CLI mode
  tui       Start a timer in TUI mode
  config    Manage plans
  stats     Show statistics
  version   Print version

Examples:
  temgo start -P classic
  temgo tui
  temgo stats --today

Use "temgo <command> -h" for more information.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(mainUsage)
		os.Exit(1)
	}

	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println("temgo", version)
		return
	}

	if err := plan.CreateTemgoDir(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create .temgo dir: %v", err)
		os.Exit(1)
	}

	if err := plan.EnsureDefaultPlans(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ensure embedded plans: %v", err)
	}

	if cfg, err := plan.LoadConfig(); err == nil {
		plan.DefaultPlanName = cfg.DefaultPlan
	}

	switch os.Args[1] {
	case "start":
		if err := commands.Start(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
			os.Exit(1)
		}
	case "tui":
		commands.StartTUI()
	case "config":
		if err := commands.RunConfig(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
			os.Exit(1)
		}
	case "stats":
		if err := commands.RunStats(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
			os.Exit(1)
		}
	}
}
