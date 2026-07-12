// Package plan defines the data model for work-timer plans (Pomodoro-style),
// provides JSON loading/saving, an embedded set of default plans,
// a custom Duration type with human-readable marshaling,
// a Builder API for programmatic plan construction, and a PlanIterator.
package plan

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Phase describes a single timed segment.
type Phase struct {
	Type     string   `json:"type"`
	Duration Duration `json:"duration"`
	Name     string   `json:"name"`
	Icon     string   `json:"icon"`
	Text     string   `json:"text"`
	Message  string   `json:"message"`
	Color    string   `json:"color"`
}

// Section groups one or more Phases and specifies how many times to repeat them.
type Section struct {
	Phases []Phase `json:"phases"`
	Repeat int     `json:"repeat"`
}

// Builder provides a fluent API for constructing a Plan in code.
type Builder struct {
	sections []Section
	repeat   int
}

// NewBuilder returns a new Builder.
func NewBuilder() *Builder {
	return &Builder{repeat: 1}
}

// AddPhase appends a single-phase section to the plan.
func (b *Builder) AddPhase(phaseType string, duration time.Duration, name, icon, text, message, color string) *Builder {
	b.sections = append(b.sections, Section{
		Phases: []Phase{{phaseType, Duration(duration), name, icon, text, message, color}},
		Repeat: 1,
	})
	return b
}

// AddRepeating appends a section with multiple phases that repeats the given number of times.
func (b *Builder) AddRepeating(repeat int, phases ...Phase) *Builder {
	b.sections = append(b.sections, Section{
		Phases: phases,
		Repeat: repeat,
	})
	return b
}

// AddSection appends a pre-built Section.
func (b *Builder) AddSection(section Section) *Builder {
	b.sections = append(b.sections, section)
	return b
}

// RepeatPlan sets how many times the entire plan cycles.
func (b *Builder) RepeatPlan(repeat int) *Builder {
	b.repeat = repeat
	return b
}

// Build finalizes the builder and returns a Plan.
func (b *Builder) Build() *Plan {
	if b.repeat < 1 {
		b.repeat = 1
	}

	return &Plan{
		Sections: b.sections,
		Repeat:   b.repeat,
	}
}

// Plan is a complete timer plan consisting of sections and a top-level repeat count.
type Plan struct {
	Name     string
	Sections []Section `json:"sections"`
	Repeat   int       `json:"repeat"`
}

// Validate checks that the plan has at least one section and a positive repeat count.
func (p *Plan) Validate() error {
	if len(p.Sections) == 0 {
		return errors.New("plan has no sections")
	}
	if p.Repeat < 1 {
		return errors.New("repeat must be >= 1")
	}
	return nil
}

// PhasesPerCycle returns the total number of phases in one full plan cycle.
func (p *Plan) PhasesPerCycle() int {
	total := 0
	for _, s := range p.Sections {
		total += len(s.Phases) * s.Repeat
	}
	return total
}

func (p Plan) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Plan: %s\n\n", p.Name))
	sb.WriteString(fmt.Sprintf("Sprints: %d\n", p.Repeat))
	for i, section := range p.Sections {
		sectionString := fmt.Sprintf("Section %d (%d×):\n", i+1, section.Repeat)
		sb.WriteString(sectionString)
		for _, phase := range section.Phases {
			phaseString := fmt.Sprintf("%s: %s\n", phase.Name, phase.Duration)
			sb.WriteString(phaseString)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
