package config

import (
	"time"
	"flag"
	"fmt"
	"io"
)

type TimerParams struct {
	Prolog time.Duration
	Work time.Duration
	Rest time.Duration
	LongRest time.Duration
	Cycles int
	Sprints int
}

var presets = map[string]TimerParams {
	"classic": {10 * time.Second, 25 * time.Minute, 5 * time.Minute, 30 * time.Minute, 4, 3},
	"short": {10 * time.Second, 15 * time.Minute, 3 * time.Minute, 15 * time.Minute, 3, 2},
	"long": {10 * time.Second, 50 * time.Minute, 10 * time.Minute, 30 * time.Minute, 3, 2},
}

func ParseFlags(args []string) (TimerParams, error) {
	fs := flag.NewFlagSet("temgo", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	presetName := fs.String("P", "", "preset name")
	prolog := fs.Duration("p", 10 * time.Second, "prologue time")
	work := fs.Duration("w", 25 * time.Minute, "work time")
	rest := fs.Duration("r",  5 * time.Minute, "rest time")
	longRest := fs.Duration("lr",  30 * time.Minute, "long rest time")
	cycles := fs.Int("c",  4, "cycles num")
	sprints := fs.Int("s",  3, "sprints num")
	
	if err := fs.Parse(args); err != nil {
		return TimerParams{}, err
	}

	if *presetName != "" {
		p, ok := presets[*presetName]
		if !ok {
			return TimerParams{}, fmt.Errorf("unknown preset: %s (use: classic, short, long)", *presetName)
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
		return TimerParams{}, fmt.Errorf("durations must be positive")
	}
	if *cycles <= 0 || *sprints <= 0 {
		return TimerParams{}, fmt.Errorf("counts must be positive")
	}
	if *prolog < 0 {
		return TimerParams{}, fmt.Errorf("prolog must be non-negative")
	}

	return TimerParams{
		Prolog: *prolog,
		Work: *work,
		Rest: *rest, 
		LongRest: *longRest,
		Cycles:*cycles, 
		Sprints: *sprints,
		}, nil
}