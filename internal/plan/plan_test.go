package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeLoadPlan(t *testing.T, json string) *Plan {
	t.Helper()
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPlan(path)
	if err != nil {
		t.Fatalf("LoadPlan: %v", err)
	}
	return p
}

const twoByTwoPlan = `{
  "sections": [
    {
      "phases": [
        {"type": "a", "duration": "1s", "name": "A", "icon": "•", "text": "", "message": "", "color": "#FFF"},
        {"type": "b", "duration": "1s", "name": "B", "icon": "•", "text": "", "message": "", "color": "#FFF"}
      ],
      "repeat": 2
    },
    {
      "phases": [
        {"type": "c", "duration": "1s", "name": "C", "icon": "•", "text": "", "message": "", "color": "#FFF"}
      ],
      "repeat": 1
    }
  ],
  "repeat": 2
}`

func TestPhasesPerCycle(t *testing.T) {
	plan := writeLoadPlan(t, twoByTwoPlan)
	if got := plan.PhasesPerCycle(); got != 5 {
		t.Errorf("PhasesPerCycle() = %d, want 5", got)
	}
}

func TestPlanIterator_PhaseCount(t *testing.T) {
	plan := writeLoadPlan(t, twoByTwoPlan)
	iter := NewPlanIterator(plan)

	phasesPerCycle := plan.PhasesPerCycle()
	totalPhases := phasesPerCycle * plan.Repeat

	count := 0
	for {
		_, ok := iter.Next()
		if !ok {
			break
		}
		count++
	}

	if count != totalPhases {
		t.Errorf("total phases = %d, want %d", count, totalPhases)
	}
}

func TestPlanIterator_PhaseOrder(t *testing.T) {
	plan := writeLoadPlan(t, twoByTwoPlan)
	iter := NewPlanIterator(plan)

	wantTypes := []string{"a", "b", "a", "b", "c"}

	for cycle := 0; cycle < plan.Repeat; cycle++ {
		for _, want := range wantTypes {
			phase, ok := iter.Next()
			if !ok {
				t.Fatalf("cycle %d: expected %q, but iterator ended", cycle+1, want)
			}
			if phase.Type != want {
				t.Errorf("cycle %d: got %q, want %q", cycle+1, phase.Type, want)
			}
		}
	}

	if _, ok := iter.Next(); ok {
		t.Error("expected iterator to end after all cycles")
	}
}

func TestPlanIterator_End(t *testing.T) {
	iter := NewPlanIterator(writeLoadPlan(t, twoByTwoPlan))

	for {
		_, ok := iter.Next()
		if !ok {
			break
		}
	}

	_, ok := iter.Next()
	if ok {
		t.Error("Next() after end should return false")
	}
	_, ok = iter.Next()
	if ok {
		t.Error("Next() after end should consistently return false")
	}
}

func TestPlanIterator_CurrentRepeat(t *testing.T) {
	plan := writeLoadPlan(t, twoByTwoPlan)
	iter := NewPlanIterator(plan)

	if got := iter.CurrentRepeat(); got != 0 {
		t.Errorf("start: CurrentRepeat() = %d, want 0", got)
	}

	for i := 0; i < plan.PhasesPerCycle(); i++ {
		iter.Next()
	}
	if got := iter.CurrentRepeat(); got != 1 {
		t.Errorf("after cycle 1: CurrentRepeat() = %d, want 1", got)
	}
}

func TestPlanIterator_CurrentRepeat_ClampedAtEnd(t *testing.T) {
	plan := writeLoadPlan(t, twoByTwoPlan)
	iter := NewPlanIterator(plan)

	for {
		_, ok := iter.Next()
		if !ok {
			break
		}
	}

	if got := iter.CurrentRepeat(); got != plan.Repeat-1 {
		t.Errorf("after end: CurrentRepeat() = %d, want %d (clamped)", got, plan.Repeat-1)
	}
}

func TestPlanIterator_Reset(t *testing.T) {
	iter := NewPlanIterator(writeLoadPlan(t, twoByTwoPlan))

	for i := 0; i < 3; i++ {
		iter.Next()
	}

	iter.Reset()

	phase, ok := iter.Next()
	if !ok {
		t.Fatal("expected phase after Reset")
	}
	if phase.Type != "a" {
		t.Errorf("after Reset: got %q, want %q", phase.Type, "a")
	}
	if iter.CurrentRepeat() != 0 {
		t.Errorf("after Reset: CurrentRepeat() = %d, want 0", iter.CurrentRepeat())
	}
}

func TestPlanIterator_SinglePhasePlan(t *testing.T) {
	json := `{
	"sections": [
		{
		"phases": [
			{"type": "simple", "duration": "1s", "name": "S", "icon": "•", "text": "", "message": "", "color": "#FFF"}
		],
		"repeat": 1
		}
	],
	"repeat": 1
	}`
	plan := writeLoadPlan(t, json)
	iter := NewPlanIterator(plan)

	phase, ok := iter.Next()
	if !ok {
		t.Fatal("expected phase")
	}
	if phase.Type != "simple" {
		t.Errorf("got %q, want %q", phase.Type, "simple")
	}

	_, ok = iter.Next()
	if ok {
		t.Error("expected end after single phase")
	}
}

func TestBuilder_RepeatPlanZero(t *testing.T) {
	plan := NewBuilder().
		AddPhase("x", time.Second, "X", "•", "", "", "#FFF").
		RepeatPlan(0).
		Build()

	if plan.Repeat != 1 {
		t.Errorf("RepeatPlan(0) should default to 1, got %d", plan.Repeat)
	}
}

func TestBuilder_RepeatPlanNegative(t *testing.T) {
	plan := NewBuilder().
		AddPhase("x", time.Second, "X", "•", "", "", "#FFF").
		RepeatPlan(-5).
		Build()

	if plan.Repeat != 1 {
		t.Errorf("RepeatPlan(-5) should default to 1, got %d", plan.Repeat)
	}
}

func TestLoadPlan(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	json := `{
	"sections": [
		{
		"phases": [
		{"type": "work", "duration": "25m", "name": "Work", "icon": "🧠", "text": "Focus", "message": "Focus!", "color": "#00FF00"},
		{"type": "rest", "duration": "5m",  "name": "Rest", "icon": "☕", "text": "Break", "message": "Break!", "color": "#87CEEB"}
		],
		"repeat": 4
		}
	],
	"repeat": 3
	}`

	if err := os.WriteFile(path, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}

	plan, err := LoadPlan(path)
	if err != nil {
		t.Fatalf("LoadPlan: %v", err)
	}

	work := plan.Sections[0].Phases[0]
	if time.Duration(work.Duration) != 25*time.Minute {
		t.Errorf("work.Duration = %v, want 25m", time.Duration(work.Duration))
	}
	rest := plan.Sections[0].Phases[1]
	if time.Duration(rest.Duration) != 5*time.Minute {
		t.Errorf("rest.Duration = %v, want 5m", time.Duration(rest.Duration))
	}
}

func TestLoadPlan_NotFound(t *testing.T) {
	_, err := LoadPlan("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadPlan_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.json"), []byte("not json"), 0644)

	_, err := LoadPlan(filepath.Join(dir, "bad.json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadPlan_BadDuration(t *testing.T) {
	dir := t.TempDir()
	json := `{
  "sections": [
    {
      "phases": [
        {"type": "x", "duration": "1s", "name": "X", "icon": "•", "text": "", "message": "", "color": "#FFF"}
      ],
      "repeat": 1
    }
  ],
  "repeat": 1,
  "bad": "x"
}`
	os.WriteFile(filepath.Join(dir, "bad.json"), []byte(json), 0644)

	plan, err := LoadPlan(filepath.Join(dir, "bad.json"))
	if err != nil {
		t.Fatalf("unexpected error for unknown field: %v", err)
	}
	if plan.Repeat != 1 {
		t.Errorf("Repeat = %d, want 1", plan.Repeat)
	}
}

func TestPhase_AllFields(t *testing.T) {
	json := `{
	"sections": [
		{
		"phases": [
			{"type": "prolog", "duration": "10s", "name": "Prolog", "icon": "🚀", "text": "Go", "message": "Go!", "color": "#00CED1"}
		],
		"repeat": 1
		}
	],
	"repeat": 1
	}`
	plan := writeLoadPlan(t, json)
	iter := NewPlanIterator(plan)

	phase, ok := iter.Next()
	if !ok {
		t.Fatal("expected phase")
	}
	if phase.Name == "" {
		t.Error("phase.Name is empty")
	}
	if phase.Icon == "" {
		t.Error("phase.Icon is empty")
	}
	if phase.Color == "" {
		t.Error("phase.Color is empty")
	}
	if phase.Text == "" {
		t.Error("phase.Text is empty")
	}
	if phase.Message == "" {
		t.Error("phase.Message is empty")
	}
}

func TestDuration_String(t *testing.T) {
	tests := []struct {
		name string
		d    Duration
		want string
	}{
		{"zero", Duration(0), "00:00"},
		{"seconds only", Duration(45 * time.Second), "00:45"},
		{"one minute", Duration(60 * time.Second), "01:00"},
		{"minutes and seconds", Duration(25*time.Minute + 45*time.Second), "25:45"},
		{"one hour", Duration(1 * time.Hour), "1:00:00"},
		{"hours minutes seconds", Duration(3*time.Hour + 35*time.Minute + 15*time.Second), "3:35:15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("Duration(%v).String() = %q, want %q", time.Duration(tt.d), got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{"zero", 0, "00:00"},
		{"seconds", 45 * time.Second, "00:45"},
		{"minute", 60 * time.Second, "01:00"},
		{"minutes seconds", 25*time.Minute + 45*time.Second, "25:45"},
		{"hour", 1 * time.Hour, "1:00:00"},
		{"hours minutes", 2*time.Hour + 15*time.Minute, "2:15:00"},
		{"full", 3*time.Hour + 35*time.Minute + 15*time.Second, "3:35:15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDuration(tt.input); got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadPlan_SetsName(t *testing.T) {
	dir := t.TempDir()
	json := `{
	"sections": [{"phases": [
		{"type": "w", "duration": "1s", "name": "W", "icon": "•", "text": "", "message": "", "color": "#FFF"}
	], "repeat": 1}], "repeat": 1}`
	os.WriteFile(filepath.Join(dir, "my-awesome-plan.json"), []byte(json), 0644)

	p, err := LoadPlan(filepath.Join(dir, "my-awesome-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "my-awesome-plan" {
		t.Errorf("Name = %q, want %q", p.Name, "my-awesome-plan")
	}
}

func TestLoadPlan_NameWithoutExt(t *testing.T) {
	dir := t.TempDir()
	json := `{"sections": [{"phases": [
		{"type": "w", "duration": "1s", "name": "W", "icon": "•", "text": "", "message": "", "color": "#FFF"}
	], "repeat": 1}], "repeat": 1}`
	os.WriteFile(filepath.Join(dir, "deep.work.json"), []byte(json), 0644)

	p, err := LoadPlan(filepath.Join(dir, "deep.work.json"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "deep.work" {
		t.Errorf("Name = %q, want %q", p.Name, "deep.work")
	}
}

func TestPlan_String(t *testing.T) {
	p := NewBuilder().
		AddPhase("work", time.Second, "Work", "🧠", "Focus", "Focus!", "#FFF").
		Build()
	p.Name = "test"

	s := p.String()

	checks := []string{
		"Plan: test",
		"Work",
		"00:01",
	}
	for _, want := range checks {
		if !strings.Contains(s, want) {
			t.Errorf("String() should contain %q\nGot:\n%s", want, s)
		}
	}
}

func TestPlan_String_MultiSection(t *testing.T) {
	p := NewBuilder().
		AddPhase("work", 25*time.Minute, "Work", "🧠", "", "", "#FFF").
		AddRepeating(2,
			Phase{Type: "rest", Duration: Duration(5 * time.Minute), Name: "Rest", Icon: "☕", Color: "#87CEEB"},
		).
		Build()
	p.Name = "multi"

	s := p.String()
	if !strings.Contains(s, "25:00") {
		t.Error("String() should contain formatted work duration")
	}
	if !strings.Contains(s, "05:00") {
		t.Error("String() should contain formatted rest duration")
	}
	if !strings.Contains(s, "(2×)") {
		t.Error("String() should show section repeat count")
	}
}
