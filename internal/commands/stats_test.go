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

func TestRunStats_Success(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"default (all)", []string{}},
		{"today", []string{"--today"}},
		{"week", []string{"--week"}},
		{"all explicit", []string{"--all"}},
		{"json export", []string{"--all", "--json"}},
		{"csv export", []string{"--all", "--csv"}},
		{"help", []string{"--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupStatsTest(t)
			err := RunStats(tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunStats_Errors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"mutually exclusive ranges", []string{"--today", "--week"}},
		{"mutually exclusive formats", []string{"--json", "--csv"}},
		{"unknown flag", []string{"--xyz"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupStatsTest(t)
			err := RunStats(tt.args)
			if err == nil {
				t.Error("expected error")
			}
		})
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
