package plan

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Phase struct {
	Type     string   `json:"type"`
	Duration Duration `json:"duration"`
	Name     string   `json:"name"`
	Icon     string   `json:"icon"`
	Text     string   `json:"text"`
	Message  string   `json:"message"`
	Color    string   `json:"color"`
}

type Section struct {
	Phases []Phase `json:"phases"`
	Repeat int     `json:"repeat"`
}

type Builder struct {
	sections []Section
	repeat   int
}

func NewBuilder() *Builder {
	return &Builder{repeat: 1}
}

func (b *Builder) AddPhase(phaseType string, duration time.Duration, name, icon, text, message, color string) *Builder {
	b.sections = append(b.sections, Section{
		Phases: []Phase{{phaseType, Duration(duration), name, icon, text, message, color}},
		Repeat: 1,
	})
	return b
}

func (b *Builder) AddRepeating(repeat int, phases ...Phase) *Builder {
	b.sections = append(b.sections, Section{
		Phases: phases,
		Repeat: repeat,
	})
	return b
}

func (b *Builder) AddSection(section Section) *Builder {
	b.sections = append(b.sections, section)
	return b
}

func (b *Builder) RepeatPlan(repeat int) *Builder {
	b.repeat = repeat
	return b
}

func (b *Builder) Build() *Plan {
	if b.repeat < 1 {
		b.repeat = 1
	}

	return &Plan{
		Sections: b.sections,
		Repeat:   b.repeat,
	}
}

type Plan struct {
	Name     string
	Sections []Section `json:"sections"`
	Repeat   int       `json:"repeat"`
}

func (p *Plan) Validate() error {
	if len(p.Sections) == 0 {
		return errors.New("plan has no sections")
	}
	if p.Repeat < 1 {
		return errors.New("repeat must be >= 1")
	}
	return nil
}

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
