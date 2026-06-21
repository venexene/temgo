package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
		t.Fatalf("expected 2 lines after 2 flushes, got %d", len(lines))
	}

	var e1, e2 Entry
	json.Unmarshal([]byte(lines[0]), &e1)
	json.Unmarshal([]byte(lines[1]), &e2)

	if e1.Type != "work" || e2.Type != "rest" {
		t.Errorf("wrong order: %q, %q", e1.Type, e2.Type)
	}
}
