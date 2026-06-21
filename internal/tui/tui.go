package tui

import (
	"context"
	"time"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/plan"
	"github.com/venexene/temgo/internal/timer"
)

type Model struct {
	plan     *plan.Plan
	iterator *plan.PlanIterator

	currentPhase plan.Phase
	remaining    time.Duration
	phaseStart   time.Time

	paused bool
	width  int
	height int

	history    *history.History
	tickCancel context.CancelFunc
}

func NewModel(plan *plan.Plan, iterator *plan.PlanIterator, history *history.History) *Model {
	return &Model{
		plan:     plan,
		iterator: iterator,
		history:  history,
	}
}

type initMsg struct{}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		return initMsg{}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initMsg:
		phase, ok := m.iterator.Next()
		if !ok {
			return m, tea.Quit
		}
		m.currentPhase = phase
		m.remaining = phase.Duration
		m.phaseStart = time.Now()
		m.paused = true
		return m, nil

	case tickMsg:
		if m.paused {
			return m, nil
		}
		m.remaining -= time.Second
		if m.remaining <= 0 {
			return m.switchPhase(true)
		}
		return m, m.tick()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeySpace:
			m.paused = !m.paused
			if !m.paused {
				return m, m.tick()
			}
			return m, nil
		case tea.KeyCtrlC:
			if err := m.history.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
			}
			return m, tea.Quit
		default:
			switch msg.String() {
			case "q":
				if err := m.history.Flush(); err != nil {
					fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
				}
				return m, tea.Quit
			case "s":
				m.remaining = 0
				return m.switchPhase(false)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

var (
	phaseStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	timerStyle  = lipgloss.NewStyle().Bold(true)
	pausedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

func (m Model) View() string {
	s := ""
	s += phaseStyle.Render(m.currentPhase.Type)
	s += "\n\n"
	s += timerStyle.Render(timer.FormatDuration(m.remaining))
	s += "\n\n"
	if m.paused {
		s += pausedStyle.Render("[PAUSED]") + "\n\n"
	}
	s += hintStyle.Render("[q] quit  [space] pause  [s] skip")
	return s
}

func (m Model) switchPhase(finished bool) (tea.Model, tea.Cmd) {
	m.history.Add(history.Entry{
		Type:     m.currentPhase.Type,
		Start:    m.phaseStart,
		Duration: int(time.Since(m.phaseStart).Seconds()),
		Finished: finished,
	})

	newPhase, ok := m.iterator.Next()
	if !ok {
		return m, tea.Quit
	}

	m.currentPhase = newPhase
	m.remaining = newPhase.Duration
	m.phaseStart = time.Now()
	return m, m.tick()
}

type tickMsg time.Time

func (m *Model) tick() tea.Cmd {
	if m.tickCancel != nil {
		m.tickCancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.tickCancel = cancel

	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			return tickMsg(time.Now())
		}
	}
}
