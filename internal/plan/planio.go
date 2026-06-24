package plan

import (
	"fmt"
	"os"
	"path/filepath"
)

func AddPlanToFolder(filename string) error {
	if _, err := LoadPlan(filename); err != nil {
		return fmt.Errorf("invalid plan: %v", err)
	}

	if err := os.MkdirAll(".temgo/plans", 0755); err != nil {
		return fmt.Errorf("creating plans dir: %v", err)
	}

	dest := filepath.Join(".temgo/plans", filepath.Base(filename))
	if err := os.Rename(filename, dest); err != nil {
		return fmt.Errorf("moving plan: %v", err)
	}

	return nil
}
