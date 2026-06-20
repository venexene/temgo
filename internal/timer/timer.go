package timer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/venexene/temgo/internal/history"
)


type WorkTimer struct {
	prolog time.Duration
	work time.Duration
	rest time.Duration
	longRest time.Duration
	cycles int
	sprints int
	history *history.History
}

func NewWorkTimer(prolog, work, rest, longRest time.Duration, cycles, sprints int) *WorkTimer {
	return &WorkTimer{
		prolog: prolog,
		work: work,
		rest: rest,
		longRest: longRest,
		cycles: cycles,
		sprints: sprints,
		history: history.NewHistory(".temgo/history.jsonl"),
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
	for s := t.sprints; s > 0; s-- {
		fmt.Println("\nGet Ready!")
		start := time.Now()
		deadline := start.Add(t.prolog)
		err := t.runPhase(ctx, deadline)
		t.history.Add(history.Entry{
			Type: "prolog", 
			Start: start,
			Duration: int(time.Since(start).Seconds()),
			Finished:err == nil,
		})
		if err != nil {
			fmt.Println("\nInterrupted")
			return
		}

		fmt.Printf("\nSprint started! Sprint %d/%d\n", t.sprints - s + 1, t.sprints)
		for c := t.cycles; c > 0; c-- {
			fmt.Printf("\nWork started! Cycle %d/%d\n", t.cycles - c + 1, t.cycles)
			start = time.Now()
			deadline = start.Add(t.work)
			err = t.runPhase(ctx, deadline)
			t.history.Add(history.Entry{
				Type: "work", 
				Start: start,
				Duration: int(time.Since(start).Seconds()),
				Finished:err == nil,
			})
			if err != nil {
				fmt.Println("\nInterrupted")
				return
			}

			fmt.Println("\nIt's time to rest!")
			start = time.Now()
			deadline = start.Add(t.rest)
			err = t.runPhase(ctx, deadline)
			t.history.Add(history.Entry{
				Type: "rest", 
				Start: start,
				Duration: int(time.Since(start).Seconds()),
				Finished:err == nil,
			})
			if err != nil {
				fmt.Println("\nInterrupted")
				return
			}
		}
		fmt.Println("\nSprint finished!")

		fmt.Println("\nIt's time for a long rest!")
		start = time.Now()
		deadline = start.Add(t.longRest)
		err = t.runPhase(ctx, deadline)
		t.history.Add(history.Entry{
			Type: "longRest", 
			Start: start,
			Duration: int(time.Since(start).Seconds()),
			Finished:err == nil,
		})
		if err != nil {
			fmt.Println("\nInterrupted")
			return
		}
		fmt.Println("Long rest finished!")
	}
	
	fmt.Println("\nAll sprints done!")
}


func formatDuration(t time.Duration) string {
	seconds := int(t.Seconds())
	if seconds >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", seconds/3600,(seconds%3600)/60, seconds%60)
	} else {
		return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
	}
}

func (t* WorkTimer) runPhase(ctx context.Context, deadline time.Time) error {
	ticker := t.startTicker(deadline, ctx)
	for remaining := range ticker {
		fmt.Printf("%s\r", formatDuration(remaining))
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
	} ()
	return res
}