package plan

import (
	"fmt"
	"os"
	"path/filepath"
)

var DataDir string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	DataDir = filepath.Join(home, ".temgo")
}

func PlansDir() string {
	return filepath.Join(DataDir, "plans")
}

func HistoryPath() string {
	return filepath.Join(DataDir, "history.jsonl")
}

func AddPlanToFolder(filename string) error {
	if _, err := LoadPlan(filename); err != nil {
		return fmt.Errorf("invalid plan: %v", err)
	}

	plansDir := PlansDir()
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		return fmt.Errorf("creating plans dir: %v", err)
	}

	dest := filepath.Join(plansDir, filepath.Base(filename))
	if err := os.Rename(filename, dest); err != nil {
		return fmt.Errorf("moving plan: %v", err)
	}

	return nil
}
