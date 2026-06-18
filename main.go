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

type Preset struct {
	Prolog time.Duration
	Work time.Duration
	Rest time.Duration
	LongRest time.Duration
	Cycles int
	Sprints int
}

var presets = map[string]Preset {
	"classic": {10 * time.Second, 25 * time.Minute, 5 * time.Minute, 30 * time.Minute, 4, 3},
	"short": {10 * time.Second, 15 * time.Minute, 3 * time.Minute, 15 * time.Minute, 3, 2},
	"long": {10 * time.Second, 50 * time.Minute, 10 * time.Minute, 30 * time.Minute, 3, 2},
}


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

	presetName := fs.String("P", "", "preset name")
	prolog := fs.Duration("p", 10 * time.Second, "prologue time")
	work := fs.Duration("w", 25 * time.Minute, "work time")
	rest := fs.Duration("r",  5 * time.Minute, "rest time")
	longRest := fs.Duration("lr",  30 * time.Minute, "long rest time")
	cycles := fs.Int("c",  4, "cycles num")
	sprints := fs.Int("s",  3, "sprints num")
	
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	if *presetName != "" {
		p, ok := presets[*presetName]
		if !ok {
			return nil, fmt.Errorf("unknown preset: %s (use: classic, short, long)", *presetName)
		}

		userFlags := make(map[string]bool)
		fs.Visit(func(f *flag.Flag) {
			userFlags[f.Name] = true
		})
		
		if !userFlags["p"]  { *prolog = p.Prolog }
		if !userFlags["w"]  { *work = p.Work }
		if !userFlags["r"]  { *rest = p.Rest }
		if !userFlags["lr"] { *longRest = p.LongRest }
		if !userFlags["c"]  { *cycles = p.Cycles }
		if !userFlags["s"]  { *sprints = p.Sprints }
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