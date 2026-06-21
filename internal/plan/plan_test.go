package plan

import (
	"testing"
	"time"
)

func TestPhasesPerCycle(t *testing.T) {
	tests := []struct {
		name string
		plan *Plan
		want int
	}{
		{"classic", ClassicPlan(), 10},
		{"short", ShortPlan(), 8},
		{"long", LongPlan(), 8},
		{"single phase", NewBuilder().
			AddPhase("only", time.Second, "Only", "•", "", "#FFF").
			Build(), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.plan.PhasesPerCycle(); got != tt.want {
				t.Errorf("PhasesPerCycle() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPlanIterator_PhaseCount(t *testing.T) {
	plan := ClassicPlan()
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
	plan := ClassicPlan()
	iter := NewPlanIterator(plan)

	wantTypes := []string{
		"prolog",
		"work", "rest", "work", "rest", "work", "rest", "work", "rest",
		"longRest",
	}

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
	iter := NewPlanIterator(ClassicPlan())

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
	plan := ClassicPlan()
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

	for i := 0; i < plan.PhasesPerCycle(); i++ {
		iter.Next()
	}
	if got := iter.CurrentRepeat(); got != 2 {
		t.Errorf("after cycle 2: CurrentRepeat() = %d, want 2", got)
	}
}

func TestPlanIterator_CurrentRepeat_ClampedAtEnd(t *testing.T) {
	plan := ClassicPlan()
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
	iter := NewPlanIterator(ClassicPlan())

	for i := 0; i < 5; i++ {
		iter.Next()
	}

	iter.Reset()

	phase, ok := iter.Next()
	if !ok {
		t.Fatal("expected phase after Reset")
	}
	if phase.Type != "prolog" {
		t.Errorf("after Reset: got %q, want %q", phase.Type, "prolog")
	}
	if iter.CurrentRepeat() != 0 {
		t.Errorf("after Reset: CurrentRepeat() = %d, want 0", iter.CurrentRepeat())
	}
}

func TestPlanIterator_SinglePhasePlan(t *testing.T) {
	plan := NewBuilder().
		AddPhase("simple", time.Second, "Simple", "•", "", "#FFF").
		Build()

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

func TestPhase_AllFields(t *testing.T) {
	plan := ClassicPlan()
	iter := NewPlanIterator(plan)

	phase, ok := iter.Next()
	if !ok {
		t.Fatal("expected prolog phase")
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
}
