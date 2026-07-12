// Package history provides an append-only JSONL journal for session entries.
package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/venexene/temgo/internal/plan"
)

// Entry represents a phase session.
type Entry struct {
	Type     string    `json:"type"`
	Start    time.Time `json:"start"`
	Duration int       `json:"duration_sec"`
	Finished bool      `json:"finished"`
}

// History holds in-memory entries and flushes them to a JSONL file.
type History struct {
	Entries  []Entry
	filePath string
}

// NewHistory creates a History that writes to the given file path.
func NewHistory(filePath string) *History {
	return &History{
		Entries:  []Entry{},
		filePath: filePath,
	}
}

// Add appends an entry to the in-memory buffer.
func (h *History) Add(entry Entry) {
	h.Entries = append(h.Entries, entry)
}

// Flush writes all buffered entries to the JSONL file and clears the buffer.
func (h *History) Flush() error {
	if err := os.MkdirAll(filepath.Dir(h.filePath), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(h.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	for _, record := range h.Entries {
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}

	h.Entries = h.Entries[:0]

	return nil
}

// LoadRange returns entries whose Start falls within [from, to] inclusive.
func LoadRange(from, to time.Time) ([]Entry, error) {
	file, err := os.Open(plan.HistoryPath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entries := []Entry{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry Entry

		if len(scanner.Bytes()) == 0 {
			continue
		}

		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			fmt.Fprintf(os.Stderr, "failed to load history entry: %v", err)
			continue
		}

		if !entry.Start.Before(from) && !entry.Start.After(to) {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// LoadAll returns every entry in the history file.
func LoadAll() ([]Entry, error) {
	return LoadRange(time.Unix(0, 0), time.Now())
}

// LoadToday returns entries from the current calendar day.
func LoadToday() ([]Entry, error) {
	now := time.Now()
	from := now.Truncate(24 * time.Hour)
	to := from.Add(24*time.Hour - time.Second)
	return LoadRange(from, to)
}

func startOfMondayBasedWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	shift := 0
	if weekday == 0 {
		shift = 6
	} else {
		shift = weekday - 1
	}
	return t.AddDate(0, 0, -shift).Truncate(24 * time.Hour)
}

// LoadWeek returns entries from the current Monday-based week.
func LoadWeek() ([]Entry, error) {
	from := startOfMondayBasedWeek(time.Now())
	to := from.Add(7*24*time.Hour - time.Second)
	return LoadRange(from, to)
}

// LoadHistory loads all entries into a new History value.
func LoadHistory() (*History, error) {
	entries, err := LoadAll()
	if err != nil {
		return nil, err
	}
	h := NewHistory(plan.HistoryPath())
	h.Entries = entries
	return h, nil
}
