package commands

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/plan"
)

func setupStatsTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	plan.CreateTemgoDir()
	plan.EnsureDefaultPlans()

	ref := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	path := plan.HistoryPath()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range []history.Entry{
		{Type: "work", Start: ref, Duration: 1500, Finished: true},
		{Type: "rest", Start: ref.Add(30 * time.Minute), Duration: 300, Finished: true},
	} {
		if err := enc.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRunStats_Default(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunStats_Today(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{"--today"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunStats_Week(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{"--week"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunStats_All(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{"--all"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunStats_MutuallyExclusive(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{"--today", "--week"})
	if err == nil {
		t.Error("expected error for --today --week")
	}
}

func TestRunStats_JSON(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{"--all", "--json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunStats_CSV(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{"--all", "--csv"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunStats_JSON_CSV_MutuallyExclusive(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{"--json", "--csv"})
	if err == nil {
		t.Error("expected error for --json --csv")
	}
}

func TestRunStats_Help(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{"--help"})
	if err != nil {
		t.Errorf("help should not return error: %v", err)
	}
}

func TestRunStats_InvalidFlag(t *testing.T) {
	setupStatsTest(t)
	err := RunStats([]string{"--xyz"})
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestRunStats_NoHistoryFile(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })
	plan.CreateTemgoDir()

	err := RunStats([]string{"--all"})
	if err != nil && !strings.Contains(err.Error(), "no such file") {
		t.Errorf("expected file not found error, got: %v", err)
	}
}
