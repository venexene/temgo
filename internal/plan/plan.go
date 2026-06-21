package plan

import "time"

type Phase struct {
	Type     string
	Duration time.Duration
	Name     string
	Icon     string
	Message  string
	Color    string
}

type Section struct {
	Phases []Phase
	Repeat int
}

type Builder struct {
	sections []Section
	repeat   int
}

func NewBuilder() *Builder {
	return &Builder{repeat: 1}
}

func (b *Builder) AddPhase(phaseType string, duration time.Duration, name, icon, message, color string) *Builder {
	b.sections = append(b.sections, Section{
		Phases: []Phase{{phaseType, duration, name, icon, message, color}},
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
	Sections []Section
	Repeat   int
}

func (p *Plan) PhasesPerCycle() int {
	total := 0
	for _, s := range p.Sections {
		total += len(s.Phases) * s.Repeat
	}
	return total
}

func ClassicPlan() *Plan {
	return NewBuilder().
		AddPhase("prolog", 10*time.Second,
			"Prolog", "🚀", "Prepare to focus", "#00CED1").
		AddRepeating(4,
			Phase{Type: "work", Duration: 25 * time.Minute,
				Name: "Work", Icon: "🧠", Message: "Stay focused", Color: "#00FF00"},
			Phase{Type: "rest", Duration: 5 * time.Minute,
				Name: "Rest", Icon: "☕", Message: "Take a break", Color: "#87CEEB"},
		).
		AddPhase("longRest", 30*time.Minute,
			"Long Rest", "😴", "Great work! Long break", "#DDA0DD").
		RepeatPlan(3).
		Build()
}

func ShortPlan() *Plan {
	return NewBuilder().
		AddPhase("prolog", 10*time.Second,
			"Prolog", "🚀", "Prepare to focus", "#00CED1").
		AddRepeating(3,
			Phase{Type: "work", Duration: 15 * time.Minute,
				Name: "Work", Icon: "🧠", Message: "Stay focused", Color: "#00FF00"},
			Phase{Type: "rest", Duration: 3 * time.Minute,
				Name: "Rest", Icon: "☕", Message: "Take a break", Color: "#87CEEB"},
		).
		AddPhase("longRest", 15*time.Minute,
			"Long Rest", "😴", "Short break session done", "#DDA0DD").
		RepeatPlan(2).
		Build()
}

func LongPlan() *Plan {
	return NewBuilder().
		AddPhase("prolog", 10*time.Second,
			"Prolog", "🚀", "Prepare for deep work", "#00CED1").
		AddRepeating(3,
			Phase{Type: "work", Duration: 50 * time.Minute,
				Name: "Deep Work", Icon: "🧠", Message: "Deep focus session", Color: "#00FF00"},
			Phase{Type: "rest", Duration: 10 * time.Minute,
				Name: "Rest", Icon: "☕", Message: "Step away, recharge", Color: "#87CEEB"},
		).
		AddPhase("longRest", 30*time.Minute,
			"Long Rest", "😴", "Outstanding! Take a real break", "#DDA0DD").
		RepeatPlan(2).
		Build()
}

type PlanIterator struct {
	plan *Plan

	planRepeat int

	sectionIndex  int
	sectionRepeat int

	phaseIndex int
}

func NewPlanIterator(plan *Plan) *PlanIterator {
	return &PlanIterator{
		plan:          plan,
		planRepeat:    0,
		sectionIndex:  0,
		sectionRepeat: 0,
		phaseIndex:    0,
	}
}

func (pi *PlanIterator) Next() (Phase, bool) {
	if pi.planRepeat >= pi.plan.Repeat {
		return Phase{}, false
	}

	section := pi.plan.Sections[pi.sectionIndex]
	phase := section.Phases[pi.phaseIndex]

	pi.phaseIndex++
	if pi.phaseIndex >= len(section.Phases) {
		pi.sectionRepeat++
		pi.phaseIndex = 0
		if pi.sectionRepeat >= section.Repeat {
			pi.sectionIndex++
			pi.sectionRepeat = 0
			if pi.sectionIndex >= len(pi.plan.Sections) {
				pi.planRepeat++
				pi.sectionIndex = 0
			}
		}
	}

	return phase, true
}

func (pi *PlanIterator) Reset() {
	pi.planRepeat = 0
	pi.sectionIndex = 0
	pi.sectionRepeat = 0
	pi.phaseIndex = 0
}

func (pi *PlanIterator) CurrentRepeat() int {
	if pi.planRepeat >= pi.plan.Repeat {
		return pi.plan.Repeat - 1
	}
	return pi.planRepeat
}

func (pi *PlanIterator) PhaseIndex() int {
	return pi.planRepeat
}
