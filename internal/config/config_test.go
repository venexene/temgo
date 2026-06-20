package config

import (
	"testing"
	"time"
)

func TestConfig_Flags(t *testing.T) {
    tests := []struct {
        name     string
        input    []string
        expected TimerParams
    }{
        {
            name:  "defaults",
            input: []string{},
            expected: TimerParams{
                Prolog:   10 * time.Second,
                Work:     25 * time.Minute,
                Rest:     5 * time.Minute,
                LongRest: 30 * time.Minute,
                Cycles:   4,
                Sprints:  3,
            },
        },
        {
            name:  "short preset",
            input: []string{"-P", "short"},
            expected: TimerParams{
                Prolog:   10 * time.Second,
                Work:     15 * time.Minute,
                Rest:     3 * time.Minute,
                LongRest: 15 * time.Minute,
                Cycles:   3,
                Sprints:  2,
            },
        },
        {
            name:  "override work",
            input: []string{"-w", "50m"},
            expected: TimerParams{
                Prolog:   10 * time.Second,
                Work:     50 * time.Minute,
                Rest:     5 * time.Minute,
                LongRest: 30 * time.Minute,
                Cycles:   4,
                Sprints:  3,
            },
        },
        {
            name:  "preset with override",
            input: []string{"-P", "short", "-w", "50m"},
            expected: TimerParams{
                Prolog:   10 * time.Second,
                Work:     50 * time.Minute,
                Rest:     3 * time.Minute,
                LongRest: 15 * time.Minute,
                Cycles:   3,
                Sprints:  2,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseFlags(tt.input)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tt.expected {
                t.Errorf("got %+v, want %+v", got, tt.expected)
            }
        })
    }
}

func TestConfig_Errors(t *testing.T) {
    tests := []struct {
        name  string
        input []string
    }{
        {"unknown preset", []string{"-P", "invalid"}},
        {"zero work", []string{"-w", "0s"}},
		{"zero rest", []string{"-r", "0s"}},
		{"zero long rest", []string{"-lr", "0s"}},
		{"negative prolog", []string{"-p", "-1s"}},
		{"negative work", []string{"-w", "-5m"}},
		{"negative rest", []string{"-r", "-5m"}},
		{"negative long rest", []string{"-lr", "-5m"}},
        {"zero sprints", []string{"-s", "0"}},
        {"zero cycles", []string{"-c", "0"}},
        {"negative sprints", []string{"-s", "-1"}},
        {"negative cycles", []string{"-c", "-1"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ParseFlags(tt.input)
            if err == nil {
                t.Error("expected error, got nil")
            }
        })
    }
}