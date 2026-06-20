package timer

import (
	"context"
	"testing"
	"time"

	"github.com/venexene/temgo/internal/plan"
)

func TestTimer_DuarationFormat(t *testing.T) {
	tests := []struct{
		name string
		input time.Duration
		expected string
	} {
		{"zero", 0, "00:00"},
		{"seconds only", 45*time.Second, "00:45"},
		{"one minute", 60*time.Second, "01:00"},
		{"minutes and seconds", 25*time.Minute + 45*time.Second, "25:45"},
		{"one hour", 60*time.Minute, "1:00:00"},
		{"hours and minutes", 2*time.Hour + 15*time.Minute, "2:15:00"},
		{"all", 3*time.Hour + 35*time.Minute + 15*time.Second, "3:35:15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.input)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTimer_Cancellation(t *testing.T) {
	wt := NewWorkTimer(plan.ClassicPlan())

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)
	go func() {
		deadline := time.Now().Add(10 * time.Second)
		errChan <- wt.runPhase(ctx, deadline)
	} ()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errChan:
		if err == nil {
			t.Error("expected error from cancelled context")
		}
	case <-time.After(2*time.Second):
		t.Fatal("timeout: runPhase didn't return after cancellation")
	}
}