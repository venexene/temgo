package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"errors"

	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/plan"
	"github.com/venexene/temgo/internal/timer"
)

const startUsage = `Usage: temgo start [flags]

Start a focused work timer in CLI mode.

Flags:
-P    plan name

Examples:
temgo start -P classic
`

func Start(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("Welcome to Temgo!")

	p, err := parseStart(args)
	if err != nil {
		fmt.Print(startUsage)
		return err
	}

	history := history.NewHistory(plan.HistoryPath())
	wt := timer.NewWorkTimer(p, history)

	if err := wt.Start(ctx); err != nil   && !errors.Is(err, context.Canceled){
		return err
	}

	fmt.Println("Bye!")

	return nil
}

func parseStart(args []string) (*plan.Plan, error) {
	for _, arg := range args {
        if arg == "-h" || arg == "--help" {
            fmt.Print(startUsage)
            return nil, nil
        }
    }

    fs := flag.NewFlagSet("temgo", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	presetName := fs.String("P", "", "plan name")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if *presetName == "" {
		*presetName = plan.DefaultPlanName
	}

	path := filepath.Join(plan.PlansDir(), *presetName+".json")
	p, err := plan.LoadPlan(path)
	if err != nil {
		return nil, fmt.Errorf("unknown preset: %s (use: classic, short, long)", *presetName)
	}

	return p, nil
}
