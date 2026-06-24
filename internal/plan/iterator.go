package plan

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
