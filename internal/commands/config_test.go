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

func TestRunConfig_SetPersists(t *testing.T) {
	setupCommandsTest(t)

	RunConfig([]string{"set", "short"})

	cfg, err := plan.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultPlan != "short" {
		t.Errorf("DefaultPlan = %q, want %q", cfg.DefaultPlan, "short")
	}
}

func TestRunConfig_SetInvalidPlan(t *testing.T) {
	setupCommandsTest(t)

	err := RunConfig([]string{"set", "nonexistent"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	cfg, _ := plan.LoadConfig()
	if cfg.DefaultPlan != "classic" {
		t.Errorf("DefaultPlan should stay classic, got %q", cfg.DefaultPlan)
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
	setupCommandsTest(t)

	err := RunConfig([]string{"show", "classic"})
	if err != nil {
		t.Errorf("show classic: %v", err)
	}
}

func TestRunConfig_ShowNotFound(t *testing.T) {
	setupCommandsTest(t)

	err := RunConfig([]string{"show", "nonexistent"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunConfig_NoArgs(t *testing.T) {
	err := RunConfig([]string{})
	if err == nil {
		t.Error("expected error for no arguments")
	}
}

func TestRunConfig_UnknownSubcommand(t *testing.T) {
	err := RunConfig([]string{"oops"})
	if err == nil {
		t.Error("expected error for unknown subcommand")
	}
}
