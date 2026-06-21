package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
	phaseNum     int

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

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.phaseNum = 1
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
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#5A56E0")).
			Padding(1, 3).
			Align(lipgloss.Center, lipgloss.Center)

	counterStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700"))
	messageStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	timerStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	pauseStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500"))
	hintKeyStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5FD7FF"))
	hintDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintSepStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	barEmptyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
)

func progressBar(elapsed, total time.Duration, width int, color lipgloss.Color) string {
	if total <= 0 {
		total = 1
	}
	ratio := float64(elapsed) / float64(total)
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}

	filled := int(ratio * float64(width))
	empty := width - filled

	fullStyle := lipgloss.NewStyle().Foreground(color)
	return fullStyle.Render(strings.Repeat("█", filled)) + barEmptyStyle.Render(strings.Repeat("░", empty))
}

func formatHints(pairs ...string) string {
	parts := make([]string, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key := pairs[i]
		desc := pairs[i+1]
		parts = append(parts, hintKeyStyle.Render(key)+" "+hintDescStyle.Render(desc))
	}
	sep := "  " + hintSepStyle.Render("│") + "  "
	return strings.Join(parts, sep)
}

func (m Model) View() string {
	totalPhases := m.plan.PhasesPerCycle()
	cycle := m.iterator.CurrentRepeat() + 1
	counter := counterStyle.Render(fmt.Sprintf("Phase %d/%d  ·  Cycle %d/%d", m.phaseNum, totalPhases, cycle, m.plan.Repeat))

	header := headerStyle.Render(fmt.Sprintf("%s %s %s", m.currentPhase.Icon, m.currentPhase.Name, m.currentPhase.Icon))

	message := messageStyle.Render(m.currentPhase.Message)

	timeStr := timerStyle.Render(timer.FormatDuration(m.remaining))

	total := m.currentPhase.Duration
	elapsed := total - m.remaining
	bar := progressBar(elapsed, total, 30, lipgloss.Color(m.currentPhase.Color))

	var pauseLine string
	if m.paused {
		pauseLine = pauseStyle.Render("⏸  PAUSED")
	}

	hints := formatHints(
		"q", "quit",
		"space", "pause",
		"s", "skip",
	)

	content := lipgloss.JoinVertical(lipgloss.Center,
		counter,
		"",
		header,
		"",
		message,
		"",
		timeStr,
		"",
		bar,
		"",
		pauseLine,
		"",
		hints,
	)

	boxWidth := 50
	if m.width > 0 && m.width-4 < boxWidth {
		boxWidth = m.width - 4
	}
	if boxWidth < 32 {
		boxWidth = 32
	}

	box := boxStyle.Width(boxWidth).Render(content)

	if m.width == 0 || m.height == 0 {
		return box
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) switchPhase(finished bool) (tea.Model, tea.Cmd) {
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

	m.phaseNum++
	if m.phaseNum > m.plan.PhasesPerCycle() {
		m.phaseNum = 1
	}

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
