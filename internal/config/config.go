package config

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/venexene/temgo/internal/plan"
)

var PlansDir = ".temgo/plans"

func ParseFlags(args []string) (*plan.Plan, error) {
	fs := flag.NewFlagSet("temgo", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	presetName := fs.String("P", "", "preset name")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if *presetName == "" {
		*presetName = "classic"
	}

	p, err := plan.LoadEmbeddedPlan(*presetName)
	if err == nil {
		return p, nil
	}

	path := filepath.Join(PlansDir, *presetName+".json")
	p, err = plan.LoadPlan(path)
	if err != nil {
		return nil, fmt.Errorf("unknown preset: %s (use: classic, short, long)", *presetName)
	}

	return p, nil
}
