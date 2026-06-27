package plan

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func CreateTemgoDir() error {
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create .temgo dir: %w", err)
	}

	if err := os.MkdirAll(PlansDir(), 0755); err != nil {
		return fmt.Errorf("failed to create plans dir: %w", err)
	}
	return nil
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

func DeletePlanFromFolder(filename string) error {
	filePath := filepath.Join(PlansDir(), filepath.Base(filename))
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("deleting plan: %v", err)
	}
	return nil
}

func LoadPlan(path string) (*Plan, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var plan Plan
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&plan)
	if err != nil {
		return nil, err
	}

	if err := plan.Validate(); err != nil {
		return nil, err
	}

	plan.Name = strings.TrimSuffix(filepath.Base(path), ".json")

	return &plan, nil
}

func ListPlanNames() ([]string, error) {
	entries, err := os.ReadDir(PlansDir())
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".json"))
	}

	return names, nil
}

//go:embed plans/*.json
var embeddedPlans embed.FS

func EnsureDefaultPlans() error {
	if err := os.MkdirAll(PlansDir(), 0755); err != nil {
		return fmt.Errorf("failed to create plans dir: %w", err)
	}

	entries, err := embeddedPlans.ReadDir("plans")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		oldPath := filepath.Join("plans", entry.Name())

		data, err := embeddedPlans.ReadFile(oldPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read embedded file: %v", err)
			continue
		}

		newPath := filepath.Join(PlansDir(), entry.Name())
		if _, err := os.Stat(newPath); err == nil {
			continue
		}

		if err := os.WriteFile(newPath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

var DefaultPlanName string

type Config struct {
	DefaultPlan string `json:"default_plan"`
}

func LoadConfig() (Config, error) {
	path := filepath.Join(DataDir, "config.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Config{DefaultPlan: "classic"}, nil
	}
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{DefaultPlan: "classic"}, nil
	}
	if cfg.DefaultPlan == "" {
		cfg.DefaultPlan = "classic"
	}

	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path := filepath.Join(DataDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
