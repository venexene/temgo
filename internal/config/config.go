package config

import (
	"flag"
	"fmt"
	"io"

	"github.com/venexene/temgo/internal/plan"
)

var presets = map[string]*plan.Plan{
	"classic": plan.ClassicPlan(),
	"short":   plan.ShortPlan(),
	"long":    plan.LongPlan(),
}

func ParseFlags(args []string) (*plan.Plan, error) {
	fs := flag.NewFlagSet("temgo", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	presetName := fs.String("P", "", "preset name")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if *presetName == "" {
		return presets["classic"], nil
	}

	p, ok := presets[*presetName]
	if !ok {
		return nil, fmt.Errorf("unknown preset: %s (use: classic, short, long)", *presetName)
	}

	return p, nil
}
