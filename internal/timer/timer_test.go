package timer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/plan"
)

func loadTestPlan(t *testing.T) *plan.Plan {
	t.Helper()
	json := `{
	"sections": [
		{
		"phases": [
			{"type": "work", "duration": "1s", "name": "W", "icon": "•", "text": "", "message": "", "color": "#FFF"}
		],
		"repeat": 1
		}
	],
	"repeat": 1
	}`
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := plan.LoadPlan(path)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestTimer_Cancellation(t *testing.T) {
	dir := t.TempDir()
	wt := NewWorkTimer(loadTestPlan(t), history.NewHistory(filepath.Join(dir, "test.jsonl")))

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)
	go func() {
		deadline := time.Now().Add(10 * time.Second)
		errChan <- wt.runPhase(ctx, deadline)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errChan:
		if err == nil {
			t.Error("expected error from canceled context")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: runPhase didn't return after cancellation")
	}
}

func TestWorkTimer_startTicker(t *testing.T) {
	wt := NewWorkTimer(loadTestPlan(t), history.NewHistory(filepath.Join(t.TempDir(), "test.jsonl")))

	ctx := context.Background()
	deadline := time.Now().Add(200 * time.Millisecond)

	ch := wt.startTicker(deadline, ctx)

	var values []time.Duration
	for v := range ch {
		values = append(values, v)
	}

	if len(values) == 0 {
		t.Error("ticker produced no values")
	}

	for i := 1; i < len(values); i++ {
		if values[i] > values[i-1] {
			t.Errorf("ticker values not decreasing: %v > %v", values[i], values[i-1])
		}
	}

	last := values[len(values)-1]
	if last > 200*time.Millisecond {
		t.Errorf("last ticker value too large: %v", last)
	}
}

func TestWorkTimer_run_ctxBetweenPhases(t *testing.T) {
	p := plan.NewBuilder().
		AddPhase("p1", 5*time.Second, "P1", "•", "", "", "#FFF").
		AddPhase("p2", 5*time.Second, "P2", "•", "", "", "#FFF").
		Build()

	wt := NewWorkTimer(p, history.NewHistory(filepath.Join(t.TempDir(), "test.jsonl")))

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		wt.run(ctx)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("run() didn't exit after ctx cancel — may have started phase 2")
	}
}

func TestWorkTimer_Start(t *testing.T) {
	p := plan.NewBuilder().
		AddPhase("work", 100*time.Millisecond, "W", "•", "", "", "#FFF").
		Build()

	wt := NewWorkTimer(p, history.NewHistory(filepath.Join(t.TempDir(), "test.jsonl")))

	if err := wt.Start(context.Background()); err != nil {
		t.Errorf("Start: %v", err)
	}
}
