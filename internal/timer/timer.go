// Package timer provides a context-aware work timer that iterates
// through plan phases, shows a countdown, records history entries,
// and triggers desktop notifications on phase transitions.
package timer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gen2brain/beeep"

	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/plan"
)

// WorkTimer runs a plan phase by phase, ticking a countdown per phase.
type WorkTimer struct {
	plan    *plan.Plan
	history *history.History
}

// NewWorkTimer creates a WorkTimer for the given plan and history store.
func NewWorkTimer(plan *plan.Plan, history *history.History) *WorkTimer {
	return &WorkTimer{
		plan:    plan,
		history: history,
	}
}

// Start begins the timer loop
func (t *WorkTimer) Start(ctx context.Context) error {
	fmt.Println("\nWorkTimer started!")
	t.run(ctx)
	if err := t.history.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "temgo: %v\n", err)
	}
	return ctx.Err()
}

func (t *WorkTimer) run(ctx context.Context) {
	iter := plan.NewPlanIterator(t.plan)
	for {
		if ctx.Err() != nil {
			return
		}

		phase, ok := iter.Next()
		if !ok {
			break
		}

		start := time.Now()
		deadline := start.Add(time.Duration(phase.Duration))
		err := t.runPhase(ctx, deadline)
		t.history.Add(history.Entry{
			Type:     phase.Type,
			Start:    start,
			Duration: int(time.Since(start).Seconds()),
			Finished: err == nil,
		})
		if err != nil {
			fmt.Println("\nInterrupted")
			return
		}
		
		if err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration); err != nil {
			fmt.Fprintf(os.Stderr, "failed to use system beep: %v", err)
		}

		if err := beeep.Notify("temgo", phase.Message, ""); err != nil {
			fmt.Fprintf(os.Stderr, "failed to use system notification: %v", err)
		}
	}
	fmt.Println("\nPlan is over!")
}

func (t *WorkTimer) runPhase(ctx context.Context, deadline time.Time) error {
	ticker := t.startTicker(deadline, ctx)
	for remaining := range ticker {
		fmt.Printf("%s\r", plan.FormatDuration(remaining))
	}
	return ctx.Err()
}

func (t *WorkTimer) startTicker(deadline time.Time, ctx context.Context) <-chan time.Duration {
	res := make(chan time.Duration)

	go func() {
		defer close(res)
		for {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return
			}

			select {
			case <-ctx.Done():
				return
			case res <- remaining:
			}

			next := remaining - time.Second
			sleepUntil := deadline.Add(-next)

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(sleepUntil)):
			}
		}
	}()
	return res
}
