package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/venexene/temgo/internal/plan"
)

func TestHistory_AddFlush(t *testing.T) {
	dir := t.TempDir()
	h := NewHistory(filepath.Join(dir, "test.jsonl"))

	testTime := time.Now().Truncate(0)
	entries := []Entry{
		{"work", testTime, 1500, true},
		{"rest", testTime, 500, false},
	}

	for _, entry := range entries {
		h.Add(entry)
	}

	if err := h.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if len(h.Entries) != 0 {
		t.Errorf("Entries not cleared after Flush, got len: %d", len(h.Entries))
	}

	data, err := os.ReadFile(filepath.Join(dir, "test.jsonl"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	for i, want := range entries {
		var got Entry

		if err := json.Unmarshal([]byte(lines[i]), &got); err != nil {
			t.Fatalf("line %d: unmarshal failed: %v", i, err)
		}

		if got.Type != want.Type {
			t.Errorf("line %d: Type = got: %q, want: %q", i, got.Type, want.Type)
		}

		if got.Duration != want.Duration {
			t.Errorf("line %d: Duration = got: %d, want: %d", i, got.Duration, want.Duration)
		}

		if got.Finished != want.Finished {
			t.Errorf("line %d: Finished = got: %v, want: %v", i, got.Finished, want.Finished)
		}

		if !got.Start.Truncate(time.Second).Equal(want.Start.Truncate(time.Second)) {
			t.Errorf("line %d: Finished = got: %v, want: %v", i, got.Start, want.Start)
		}
	}
}

func TestHistory_EmptyFlush(t *testing.T) {
	dir := t.TempDir()
	h := NewHistory(filepath.Join(dir, "test.jsonl"))

	if err := h.Flush(); err != nil {
		t.Fatalf("Flush empty history: %v", err)
	}
}

func TestHistory_MultipleFlushes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")
	h := NewHistory(path)

	h.Add(Entry{Type: "work", Duration: 100, Finished: true})
	if err := h.Flush(); err != nil {
		t.Fatalf("flush 1: %v", err)
	}

	h.Add(Entry{Type: "rest", Duration: 50, Finished: false})
	if err := h.Flush(); err != nil {
		t.Fatalf("flush 2: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	for i, want := range []Entry{
		{Type: "work", Duration: 100, Finished: true},
		{Type: "rest", Duration: 50, Finished: false},
	} {
		var got Entry
		if err := json.Unmarshal([]byte(lines[i]), &got); err != nil {
			t.Fatalf("line %d: %v", i, err)
		}
		if got.Duration != want.Duration || got.Finished != want.Finished {
			t.Errorf("line %d mismatch: got %+v", i, got)
		}
	}
}

func writeHistoryJSONL(t *testing.T, dir string, entries []Entry) {
	t.Helper()
	path := filepath.Join(dir, "history.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadRange_Normal(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	ref := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	writeHistoryJSONL(t, dir, []Entry{
		{Type: "work", Start: ref, Duration: 1500, Finished: true},
		{Type: "rest", Start: ref.Add(30 * time.Minute), Duration: 300, Finished: true},
		{Type: "work", Start: ref.Add(-24 * time.Hour), Duration: 900, Finished: false},
		{Type: "longRest", Start: ref.Add(48 * time.Hour), Duration: 1800, Finished: true},
	})

	got, err := LoadRange(ref.Add(-1*time.Hour), ref.Add(1*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2", len(got))
	}
}

func TestLoadRange_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	os.WriteFile(filepath.Join(dir, "history.jsonl"), []byte{}, 0644)

	got, err := LoadRange(time.Time{}, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %d entries, want 0", len(got))
	}
}

func TestLoadRange_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	_, err := LoadRange(time.Time{}, time.Now())
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadRange_SkipsMalformed(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	ref := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	os.WriteFile(filepath.Join(dir, "history.jsonl"), []byte(
		`{"type":"work","start":"`+ref.Format(time.RFC3339Nano)+`","duration_sec":100,"finished":true}
not json
{"type":"rest","start":"`+ref.Add(time.Hour).Format(time.RFC3339Nano)+`","duration_sec":50,"finished":false}
`), 0644)

	got, err := LoadRange(time.Time{}, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2 (skipped malformed)", len(got))
	}
}

func TestLoadRange_SkipsEmptyLines(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	ref := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	os.WriteFile(filepath.Join(dir, "history.jsonl"), []byte(
		"\n{\"type\":\"work\",\"start\":\""+ref.Format(time.RFC3339Nano)+"\",\"duration_sec\":100,\"finished\":true}\n\n"),
		0644)

	got, err := LoadRange(time.Time{}, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1 (skipped empty lines)", len(got))
	}
}

func TestLoadRange_Boundaries(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	ref := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	writeHistoryJSONL(t, dir, []Entry{
		{Type: "work", Start: ref, Duration: 100, Finished: true},
		{Type: "rest", Start: ref.Add(time.Hour), Duration: 50, Finished: true},
	})

	got, err := LoadRange(ref, ref.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2 (boundaries inclusive)", len(got))
	}
}

func TestLoadAll(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	writeHistoryJSONL(t, dir, []Entry{
		{Type: "work", Start: time.Now(), Duration: 100, Finished: true},
		{Type: "rest", Start: time.Now().Add(-48 * time.Hour), Duration: 50, Finished: false},
	})

	got, err := LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2", len(got))
	}
}

func TestLoadToday(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	writeHistoryJSONL(t, dir, []Entry{
		{Type: "work", Start: today.Add(time.Hour), Duration: 100, Finished: true},
		{Type: "rest", Start: today.Add(-24 * time.Hour), Duration: 50, Finished: false},
	})

	got, err := LoadToday()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1 (only today)", len(got))
	}
}

func TestLoadWeek(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	now := time.Now()
	thisMonday := startOfMondayBasedWeek(now)
	writeHistoryJSONL(t, dir, []Entry{
		{Type: "work", Start: thisMonday.Add(time.Hour), Duration: 100, Finished: true},
		{Type: "rest", Start: thisMonday.Add(-24 * time.Hour), Duration: 50, Finished: false},
	})

	got, err := LoadWeek()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1 (only this week)", len(got))
	}
}

func TestLoadHistory(t *testing.T) {
	dir := t.TempDir()
	old := plan.DataDir
	plan.DataDir = dir
	t.Cleanup(func() { plan.DataDir = old })

	writeHistoryJSONL(t, dir, []Entry{
		{Type: "work", Start: time.Now(), Duration: 100, Finished: true},
	})

	h, err := LoadHistory()
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(h.Entries))
	}
}
