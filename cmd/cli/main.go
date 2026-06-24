package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/venexene/temgo/internal/config"
	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/timer"
	"github.com/venexene/temgo/internal/plan"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if len(os.Args) >= 2 && os.Args[1] == "--add" {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "temgo: --add requires a file path")
			os.Exit(1)
		}
		if err := plan.AddPlanToFolder(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
		}
		return
	}

	fmt.Println("Welcome to Temgo!")

	plan, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
		os.Exit(2)
	}

	history := history.NewHistory(".temgo/history.jsonl")
	wt := timer.NewWorkTimer(plan, history)

	if err := wt.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
	}

	fmt.Println("Bye!")
}
