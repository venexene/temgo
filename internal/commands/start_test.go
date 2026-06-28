package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/venexene/temgo/internal/plan"
)

func setupStartTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	plan.DefaultPlanName = "classic"
	t.Cleanup(func() {
		plan.DataDir = old
		plan.DefaultPlanName = "classic"
	})
	plan.CreateTemgoDir()
	if err := plan.EnsureDefaultPlans(); err != nil {
		t.Fatalf("EnsureDefaultPlans: %v", err)
	}
}

func TestParseStart_EmbeddedPresets(t *testing.T) {
	setupStartTest(t)

	tests := []struct {
		args []string
	}{
		{args: nil},
		{args: []string{}},
		{args: []string{"-P", "classic"}},
		{args: []string{"-P", "short"}},
		{args: []string{"-P", "long"}},
	}

	for _, tt := range tests {
		name := strings.Join(tt.args, " ")
		if name == "" {
			name = "default"
		}
		t.Run(name, func(t *testing.T) {
			got, err := parseStart(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil plan")
			}
			if err := got.Validate(); err != nil {
				t.Errorf("plan failed validation: %v", err)
			}
			if len(got.Sections) == 0 {
				t.Error("plan has no sections")
			}
		})
	}
}

func TestParseStart_UnknownPreset(t *testing.T) {
	setupStartTest(t)

	_, err := parseStart([]string{"-P", "nonsense"})
	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
	if !strings.Contains(err.Error(), "unknown preset") {
		t.Errorf("error = %q, want to contain 'unknown preset'", err.Error())
	}
}

func TestParseStart_InvalidFlag(t *testing.T) {
	_, err := parseStart([]string{"-X"})
	if err == nil {
		t.Fatal("expected error for invalid flag")
	}
}

func TestParseStart_MissingValue(t *testing.T) {
	_, err := parseStart([]string{"-P"})
	if err == nil {
		t.Fatal("expected error for -P without value")
	}
}

func TestParseStart_DiskFallback(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "plans")
	os.MkdirAll(plansDir, 0755)

	json := `{
		"sections": [{
			"phases": [
				{"type": "work", "duration": "1s", "name": "T", "icon": "•", "text": "", "message": "", "color": "#FFF"}
			],
			"repeat": 1
		}],
		"repeat": 1
	}`
	os.WriteFile(filepath.Join(plansDir, "test.json"), []byte(json), 0644)

	old := plan.DataDir
	plan.DataDir = dir
	defer func() { plan.DataDir = old }()

	got, err := parseStart([]string{"-P", "test"})
	if err != nil {
		t.Fatalf("disk fallback failed: %v", err)
	}
	if got == nil {
		t.Fatal("got nil plan from disk")
	}
	if err := got.Validate(); err != nil {
		t.Errorf("disk plan failed validation: %v", err)
	}
}

func TestParseStart_InvalidPlanOnDisk(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "plans")
	os.MkdirAll(plansDir, 0755)

	json := `{"sections": [], "repeat": 1}`
	os.WriteFile(filepath.Join(plansDir, "bad.json"), []byte(json), 0644)

	old := plan.DataDir
	plan.DataDir = dir
	defer func() { plan.DataDir = old }()

	_, err := parseStart([]string{"-P", "bad"})
	if err == nil {
		t.Fatal("expected error for plan with no sections")
	}
}

func TestParseStart_Help(t *testing.T) {
	p, err := parseStart([]string{"-h"})
	if err != nil {
		t.Errorf("help should not return error: %v", err)
	}
	if p != nil {
		t.Error("help should return nil plan")
	}
}
