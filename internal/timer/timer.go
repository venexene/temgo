package timer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/plan"
)

type WorkTimer struct {
	plan    *plan.Plan
	history *history.History
}

func NewWorkTimer(plan *plan.Plan, history *history.History) *WorkTimer {
	return &WorkTimer{
		plan:    plan,
		history: history,
	}
}

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
		select {
		case <-ctx.Done():
			return
		default:
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
	}
	fmt.Println("\nPlan is over!")
}

func FormatDuration(t time.Duration) string {
	seconds := int(t.Seconds())
	if seconds >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", seconds/3600, (seconds%3600)/60, seconds%60)
	} else {
		return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
	}
}

func (t *WorkTimer) runPhase(ctx context.Context, deadline time.Time) error {
	ticker := t.startTicker(deadline, ctx)
	for remaining := range ticker {
		fmt.Printf("%s\r", FormatDuration(remaining))
	}
	fmt.Println("\a")
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
