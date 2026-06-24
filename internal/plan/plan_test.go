package plan

import (
	"os"
	"path/filepath"
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
        {"type": "a", "duration": "1s", "name": "A", "icon": "•", "message": "", "color": "#FFF"},
        {"type": "b", "duration": "1s", "name": "B", "icon": "•", "message": "", "color": "#FFF"}
      ],
      "repeat": 2
    },
    {
      "phases": [
        {"type": "c", "duration": "1s", "name": "C", "icon": "•", "message": "", "color": "#FFF"}
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
			{"type": "simple", "duration": "1s", "name": "S", "icon": "•", "message": "", "color": "#FFF"}
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
		AddPhase("x", time.Second, "X", "•", "", "#FFF").
		RepeatPlan(0).
		Build()

	if plan.Repeat != 1 {
		t.Errorf("RepeatPlan(0) should default to 1, got %d", plan.Repeat)
	}
}

func TestBuilder_RepeatPlanNegative(t *testing.T) {
	plan := NewBuilder().
		AddPhase("x", time.Second, "X", "•", "", "#FFF").
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
			{"type": "work", "duration": "25m", "name": "Work", "icon": "🧠", "message": "Focus", "color": "#00FF00"},
			{"type": "rest", "duration": "5m",  "name": "Rest", "icon": "☕", "message": "Break", "color": "#87CEEB"}
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
        {"type": "x", "duration": "1s", "name": "X", "icon": "•", "message": "", "color": "#FFF"}
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
			{"type": "prolog", "duration": "10s", "name": "Prolog", "icon": "🚀", "message": "Go", "color": "#00CED1"}
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
	if phase.Message == "" {
		t.Error("phase.Message is empty")
	}
}
