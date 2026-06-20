package plan

import "time"


type Phase struct {
	Type string
	Duration time.Duration
}

type Section struct {
	Phases []Phase
	Repeat int
}


type Builder struct {
	sections []Section
	repeat int
}

func NewBuilder() *Builder {
	return &Builder{repeat:1}
}

func (b *Builder) AddPhase(name string, duration time.Duration) *Builder {
	b.sections = append(b.sections, Section{
		Phases: []Phase{{name, duration}},
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
		Repeat: b.repeat,
	}
}


type Plan struct {
	Sections []Section
	Repeat int
}

func ClassicPlan() *Plan {
	return NewBuilder().
    AddPhase("prolog", 10*time.Second).
    AddRepeating(4,
        Phase{"work", 25*time.Minute},
        Phase{"rest", 5*time.Minute},
    ).
    AddPhase("longRest", 30*time.Minute).
    RepeatPlan(3).  // 3 спринта
    Build()
}

func ShortPlan() *Plan {
    return NewBuilder().
        AddPhase("prolog", 10*time.Second).
        AddRepeating(3,
            Phase{"work", 15 * time.Minute},
            Phase{"rest", 3 * time.Minute},
        ).
        AddPhase("longRest", 15*time.Minute).
        RepeatPlan(2).
        Build()
}

func LongPlan() *Plan {
    return NewBuilder().
        AddPhase("prolog", 10*time.Second).
        AddRepeating(3,
            Phase{"work", 50 * time.Minute},
            Phase{"rest", 10 * time.Minute},
        ).
        AddPhase("longRest", 30*time.Minute).
        RepeatPlan(2).
        Build()
}


type PlanIterator struct {
	plan *Plan

	planRepeat int

	sectionIndex int
	sectionRepeat int

	phaseIndex int
}

func NewPlanIterator(plan *Plan) *PlanIterator {
	return &PlanIterator{
		plan: plan,
		planRepeat: 0,
		sectionIndex: 0,
		sectionRepeat: 0,
		phaseIndex: 0,
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