package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig_NotExists(t *testing.T) {
	dir := t.TempDir()
	old := DataDir
	DataDir = dir
	t.Cleanup(func() { DataDir = old })

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultPlan != "classic" {
		t.Errorf("DefaultPlan = %q, want %q", cfg.DefaultPlan, "classic")
	}
}

func TestConfig_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	old := DataDir
	DataDir = dir
	t.Cleanup(func() { DataDir = old })

	if err := SaveConfig(Config{DefaultPlan: "short"}); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultPlan != "short" {
		t.Errorf("DefaultPlan = %q, want %q", cfg.DefaultPlan, "short")
	}
}

func TestLoadConfig_BrokenJSON(t *testing.T) {
	dir := t.TempDir()
	old := DataDir
	DataDir = dir
	t.Cleanup(func() { DataDir = old })

	os.WriteFile(filepath.Join(dir, "config.json"), []byte("not json"), 0644)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultPlan != "classic" {
		t.Errorf("DefaultPlan = %q, want %q (fallback for broken JSON)", cfg.DefaultPlan, "classic")
	}
}

func TestLoadConfig_EmptyDefaultPlan(t *testing.T) {
	dir := t.TempDir()
	old := DataDir
	DataDir = dir
	t.Cleanup(func() { DataDir = old })

	os.WriteFile(filepath.Join(dir, "config.json"),
		[]byte(`{"default_plan": ""}`), 0644)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultPlan != "classic" {
		t.Errorf("DefaultPlan = %q, want %q (fallback for empty)", cfg.DefaultPlan, "classic")
	}
}

func TestDeletePlanFromFolder(t *testing.T) {
	dir := t.TempDir()
	old := DataDir
	DataDir = dir
	t.Cleanup(func() { DataDir = old })

	os.MkdirAll(PlansDir(), 0755)
	path := filepath.Join(PlansDir(), "test.json")
	os.WriteFile(path, []byte(`{"sections": [{"phases": [
		{"type": "w", "duration": "1s", "name": "W", "icon": "•", "text": "", "message": "", "color": "#FFF"}
	], "repeat": 1}], "repeat": 1}`), 0644)

	if err := DeletePlanFromFolder("test.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestDeletePlanFromFolder_NotFound(t *testing.T) {
	dir := t.TempDir()
	old := DataDir
	DataDir = dir
	t.Cleanup(func() { DataDir = old })

	os.MkdirAll(PlansDir(), 0755)

	err := DeletePlanFromFolder("nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent plan")
	}
}

func TestEnsureDefaultPlans_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	old := DataDir
	DataDir = dir
	t.Cleanup(func() { DataDir = old })

	os.MkdirAll(PlansDir(), 0755)

	if err := EnsureDefaultPlans(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"classic", "short", "long"} {
		path := filepath.Join(PlansDir(), name+".json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("%s.json not created", name)
		}
	}
}

func TestEnsureDefaultPlans_NoOverwrite(t *testing.T) {
	dir := t.TempDir()
	old := DataDir
	DataDir = dir
	t.Cleanup(func() { DataDir = old })

	os.MkdirAll(PlansDir(), 0755)
	custom := `{"sections": [{"phases": [
		{"type": "x", "duration": "99s", "name": "Custom", "icon": "•", "text": "", "message": "", "color": "#FFF"}
	], "repeat": 1}], "repeat": 1}`
	os.WriteFile(filepath.Join(PlansDir(), "classic.json"), []byte(custom), 0644)

	if err := EnsureDefaultPlans(); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(PlansDir(), "classic.json"))
	if !strings.Contains(string(data), "99s") {
		t.Error("custom classic.json was overwritten by embedded")
	}
}

func TestEnsureDefaultPlans_CreatesMissingDirs(t *testing.T) {
	dir := t.TempDir()
	old := DataDir
	DataDir = dir
	t.Cleanup(func() { DataDir = old })

	if err := EnsureDefaultPlans(); err != nil {
		t.Fatal(err)
	}

	_, err := os.Stat(PlansDir())
	if os.IsNotExist(err) {
		t.Error("PlansDir should be created by EnsureDefaultPlans")
	}
}
