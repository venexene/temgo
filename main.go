package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/venexene/temgo/internal/timer"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("Welcome to Temgo!")

	t, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
		os.Exit(2)
	}

	if err := t.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
	}
	
	fmt.Println("Bye!")
}

func parseFlags() (*timer.WorkTimer, error) {
	fs := flag.NewFlagSet("temgo", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	prolog := fs.Duration("p", 10 * time.Second, "prologue time")
	work := fs.Duration("w", 25 * time.Minute, "work time")
	rest := fs.Duration("r",  5 * time.Minute, "rest time")
	longRest := fs.Duration("lr",  30 * time.Minute, "long rest time")
	cycles := fs.Int("c",  4, "cycles num")
	sprints := fs.Int("s",  3, "sprints num")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	if *work <= 0 || *rest <= 0 || *longRest <= 0 {
		return nil, fmt.Errorf("durations must be positive")
	}
	if *cycles <= 0 || *sprints <= 0 {
		return nil, fmt.Errorf("counts must be positive")
	}
	if *prolog < 0 {
		return nil, fmt.Errorf("prolog must be non-negative")
	}

	return timer.NewWorkTimer(*prolog, *work, *rest, *longRest, *cycles, *sprints), nil
}