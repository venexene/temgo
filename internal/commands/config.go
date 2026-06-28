package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/venexene/temgo/internal/plan"
)

const configUsage = `Usage: temgo config <subcommand> [arguments]

Manage presets and custom plans.

Subcommands:
list          List all available plans
show <name>   Show plan details
set		    Set default plan
add <file>    Add a custom JSON plan
delete <name> Delete plan

Examples:
temgo config list
temgo config add ~/myplan.json
temgo config show classic
`

func RunConfig(args []string) error {
	if len(args) < 1 {
		fmt.Print(configUsage)
		return fmt.Errorf("Not enough arguments")
	}

	switch args[0] {
	case "list":
		printPlanNames()
	case "set":
		if len(args) < 2 {
			fmt.Print(configUsage)
			return fmt.Errorf("Not enough arguments to set default plan")
		}
		setDefaultPlan(args[1])
	case "add":
		if len(args) < 2 {
			fmt.Print(configUsage)
			return fmt.Errorf("Not enough arguments to add new plan")
		}
		addNewPlan(args[1])
	case "delete":
		if len(args) < 2 {
			fmt.Print(configUsage)
			return fmt.Errorf("Not enough arguments to delete plan")
		}
		deletePlan(args[1])
	case "show":
		if len(args) < 2 {
			fmt.Print(configUsage)
			return fmt.Errorf("Not enough arguments to delete plan")
		}
		printPlanInfo(args[1])
	case "--help":
		fallthrough
	case "-h":
		fmt.Print(configUsage)
	default:
		fmt.Print(configUsage)
		return fmt.Errorf("Unknown command")
	}

	return nil
}

func printPlanNames() {
	names, err := plan.ListPlanNames()
	if err != nil {
		fmt.Printf("Failed to get plan names: %v\n", err)
		return
	}

	lines := make([]string, len(names))
	for i, name := range names {
		lines[i] = "- " + name
	}

	fmt.Println(strings.Join(lines, "\n"))
}

func setDefaultPlan(name string) {
	path := filepath.Join(plan.PlansDir(), fmt.Sprintf("%s.json", name))
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("File of %s plan doesn't exist\n", name)
		return
	}

	plan.DefaultPlanName = name
	cfg := plan.Config{DefaultPlan: name}
	if err := plan.SaveConfig(cfg); err != nil {
		fmt.Printf("Failed to save config: %v\n", err)
	}
}

func addNewPlan(filename string) {
	if err := plan.AddPlanToFolder(filename); err != nil {
		fmt.Printf("Failed to add plan: %v\n", err)
	}
}

func deletePlan(name string) {
	if err := plan.DeletePlanFromFolder(fmt.Sprintf("%s.json", name)); err != nil {
		fmt.Printf("Failed to delete plan: %v\n", err)
	}
}

func printPlanInfo(name string) {
	path := filepath.Join(plan.PlansDir(), fmt.Sprintf("%s.json", name))
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("File of %s plan doesn't exist\n", name)
		return
	}

	plan, err := plan.LoadPlan(path)
	if err != nil {
		fmt.Printf("Failed to load %s plan\n", name)
		return
	}

	fmt.Println()
	fmt.Print(plan)
}
