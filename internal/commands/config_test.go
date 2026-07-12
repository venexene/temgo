package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/venexene/temgo/internal/plan"
)

func setupCommandsTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	plan.CreateTemgoDir()
	if err := plan.EnsureDefaultPlans(); err != nil {
		t.Fatalf("EnsureDefaultPlans: %v", err)
	}
}

func TestRunConfig_List(t *testing.T) {
	setupCommandsTest(t)

	err := RunConfig([]string{"list"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunConfig_Set(t *testing.T) {
	tests := []struct {
		name     string
		planName string
		wantPlan string
	}{
		{"valid plan", "short", "short"},
		{"invalid plan keeps default", "nonexistent", "classic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupCommandsTest(t)

			RunConfig([]string{"set", tt.planName})

			cfg, _ := plan.LoadConfig()
			if cfg.DefaultPlan != tt.wantPlan {
				t.Errorf("DefaultPlan = %q, want %q", cfg.DefaultPlan, tt.wantPlan)
			}
		})
	}
}

func TestRunConfig_AddDelete(t *testing.T) {
	setupCommandsTest(t)

	srcDir := t.TempDir()
	json := `{"sections": [{"phases": [
		{"type": "w", "duration": "1s", "name": "W", "icon": "•", "text": "", "message": "", "color": "#FFF"}
	], "repeat": 1}], "repeat": 1}`
	os.WriteFile(filepath.Join(srcDir, "testplan.json"), []byte(json), 0644)

	RunConfig([]string{"add", filepath.Join(srcDir, "testplan.json")})

	names, _ := plan.ListPlanNames()
	found := false
	for _, name := range names {
		if name == "testplan" {
			found = true
			break
		}
	}
	if !found {
		t.Error("plan should be added")
	}

	RunConfig([]string{"delete", "testplan"})

	names, _ = plan.ListPlanNames()
	for _, name := range names {
		if name == "testplan" {
			t.Error("plan should be deleted")
		}
	}
}

func TestRunConfig_Show(t *testing.T) {
	tests := []struct {
		name     string
		planName string
	}{
		{"existing plan", "classic"},
		{"nonexistent plan", "nonexistent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupCommandsTest(t)

			err := RunConfig([]string{"show", tt.planName})
			if err != nil {
				t.Errorf("show %s: %v", tt.planName, err)
			}
		})
	}
}

func TestRunConfig_Errors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no arguments", []string{}},
		{"unknown subcommand", []string{"oops"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunConfig(tt.args)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}
