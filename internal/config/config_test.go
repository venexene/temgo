package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFlags_EmbeddedPresets(t *testing.T) {
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
			got, err := ParseFlags(tt.args)
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

func TestParseFlags_UnknownPreset(t *testing.T) {
	_, err := ParseFlags([]string{"-P", "nonsense"})
	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
	if !strings.Contains(err.Error(), "unknown preset") {
		t.Errorf("error = %q, want to contain 'unknown preset'", err.Error())
	}
}

func TestParseFlags_InvalidFlag(t *testing.T) {
	_, err := ParseFlags([]string{"-X"})
	if err == nil {
		t.Fatal("expected error for invalid flag")
	}
}

func TestParseFlags_MissingValue(t *testing.T) {
	_, err := ParseFlags([]string{"-P"})
	if err == nil {
		t.Fatal("expected error for -P without value")
	}
}

func TestParseFlags_DiskFallback(t *testing.T) {
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

	old := PlansDir
	PlansDir = plansDir
	defer func() { PlansDir = old }()

	got, err := ParseFlags([]string{"-P", "test"})
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

func TestParseFlags_EmbeddedWinsOverDisk(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "plans")
	os.MkdirAll(plansDir, 0755)

	json := `{"sections": [], "repeat": 1}`
	os.WriteFile(filepath.Join(plansDir, "classic.json"), []byte(json), 0644)

	old := PlansDir
	PlansDir = plansDir
	defer func() { PlansDir = old }()

	got, err := ParseFlags([]string{"-P", "classic"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("embedded plan failed validation: %v", err)
	}
	if len(got.Sections) == 0 {
		t.Error("embedded plan should have sections, disk fallback returned empty")
	}
}
