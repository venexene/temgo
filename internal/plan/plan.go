package plan

import (
	"encoding/json"
	"errors"
	"os"
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

func LoadPlan(filepath string) (*Plan, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var plan Plan
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&plan)
	if err != nil {
		return nil, err
	}

	if err := plan.Validate(); err != nil {
		return nil, err
	}

	return &plan, nil
}
