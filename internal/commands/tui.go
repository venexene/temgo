package commands

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/plan"
	"github.com/venexene/temgo/internal/tui"
)

// StartTUI launches the Bubble Tea TUI timer.
func StartTUI() {
	hist := history.NewHistory(plan.HistoryPath())

	model := tui.NewModel(nil, nil, hist)

	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
		os.Exit(1)
	}
}
