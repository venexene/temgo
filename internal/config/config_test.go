package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/venexene/temgo/internal/plan"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *plan.Plan
		wantErr bool
		errMsg  string
	}{
		{
			name: "no args returns classic",
			args: []string{},
			want: presets["classic"],
		},
		{
			name: "classic explicitly",
			args: []string{"-P", "classic"},
			want: presets["classic"],
		},
		{
			name: "short preset",
			args: []string{"-P", "short"},
			want: presets["short"],
		},
		{
			name: "long preset",
			args: []string{"-P", "long"},
			want: presets["long"],
		},
		{
			name:    "unknown preset",
			args:    []string{"-P", "unknown"},
			wantErr: true,
			errMsg:  "unknown preset: unknown (use: classic, short, long)",
		},
		{
			name:    "invalid flag",
			args:    []string{"-X", "value"},
			wantErr: true,
		},
		{
			name:    "-P without value",
			args:    []string{"-P"},
			wantErr: true,
		},
		{
			name: "extra args ignored",
			args: []string{"-P", "short", "extra"},
			want: presets["short"],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil plan")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("plan mismatch:\n got  %+v\n want %+v", got, tt.want)
			}
		})
	}
}

func TestParseFlags_ReturnsSameInstance(t *testing.T) {
	a, _ := ParseFlags([]string{"-P", "classic"})
	b, _ := ParseFlags([]string{"-P", "classic"})
	if a != b {
		t.Error("ParseFlags should return the same preset instance")
	}
}
