package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"

	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/plan"
	"github.com/venexene/temgo/internal/timer"
)

type state int

const (
	stateSelecting state = iota
	stateRunning
)

type Model struct {
	state  state
	plans  []planItem
	cursor int

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

type planItem struct {
	name string
	plan *plan.Plan
}

func (m *Model) loadPlans(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			plan, err := plan.LoadPlan(filepath.Join(dir, file.Name()))
			if err != nil {
				continue
			}
			m.plans = append(m.plans, planItem{
				name: strings.TrimSuffix(file.Name(), ".json"),
				plan: plan,
			})
		}
	}
}

func (m *Model) startPlan() (tea.Model, tea.Cmd) {
	m.plan = m.plans[m.cursor].plan
	m.cursor = 0
	m.iterator = plan.NewPlanIterator(m.plan)
	phase, ok := m.iterator.Next()
	if !ok {
		return m, tea.Quit
	}
	m.currentPhase = phase
	m.remaining = time.Duration(phase.Duration)
	m.phaseStart = time.Now()
	m.paused = true
	m.phaseNum = 1
	m.state = stateRunning
	return m, nil
}

type initMsg struct{}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		return initMsg{}
	}
}

func (m *Model) updateTimer(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
				m.state = stateSelecting
				return m, nil
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

func (m *Model) updateSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.plans)-1 {
				m.cursor++
			}
		case "enter":
			return m.startPlan()
		case "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initMsg:
		m.loadPlans("plans")
		m.loadPlans(".temgo/plans")
		m.state = stateSelecting
	default:
		if m.state == stateSelecting {
			return m.updateSelector(msg)
		}
		return m.updateTimer(msg)
	}
	return m, nil
}

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#5A56E0")).
			Padding(1, 3).
			Align(lipgloss.Center, lipgloss.Center)

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700"))
	counterStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700"))
	textStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
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

func (m Model) viewTimer() string {
	totalPhases := m.plan.PhasesPerCycle()
	cycle := m.iterator.CurrentRepeat() + 1
	counter := counterStyle.Render(fmt.Sprintf("Phase %d/%d  ·  Cycle %d/%d", m.phaseNum, totalPhases, cycle, m.plan.Repeat))

	header := headerStyle.Render(fmt.Sprintf("%s %s %s", m.currentPhase.Icon, m.currentPhase.Name, m.currentPhase.Icon))

	text := textStyle.Render(m.currentPhase.Text)

	timeStr := timerStyle.Render(timer.FormatDuration(m.remaining))

	total := time.Duration(m.currentPhase.Duration)
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
		text,
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

func (m Model) viewSelector() string {
	var lines []string
	lines = append(lines, titleStyle.Render("Select Plan"))

	for i, item := range m.plans {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		lines = append(lines, prefix+item.name)
	}

	hints := formatHints(
		"↑↓", "choose",
		"enter", "confirm",
		"q", "quit",
	)

	lines = append(lines, "", hints)

	content := lipgloss.JoinVertical(lipgloss.Center, lines...)

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

func (m Model) View() string {
	if m.state == stateSelecting {
		return m.viewSelector()
	}
	return m.viewTimer()
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
		m.state = stateSelecting
		return m, nil
	}

	beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	beeep.Notify("temgo", newPhase.Message, "")

	m.currentPhase = newPhase
	m.remaining = time.Duration(newPhase.Duration)
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
