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

type Entry struct {
	Type     string    `json:"type"`
	Start    time.Time `json:"start"`
	Duration int       `json:"duration_sec"`
	Finished bool      `json:"finished"`
}

type History struct {
	Entries  []Entry
	filePath string
}

func NewHistory(filePath string) *History {
	return &History{
		Entries:  []Entry{},
		filePath: filePath,
	}
}

func (h *History) Add(entry Entry) {
	h.Entries = append(h.Entries, entry)
}

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

func LoadAll() ([]Entry, error) {
	return LoadRange(time.Unix(0, 0), time.Now())
}

func LoadToday() ([]Entry, error) {
    now := time.Now()
    from := now.Truncate(24 * time.Hour)
    to := from.Add(24 * time.Hour - time.Second)
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

func LoadWeek() ([]Entry, error) {
    from := startOfMondayBasedWeek(time.Now())
    to := from.Add(7 * 24 * time.Hour - time.Second)
    return LoadRange(from, to)
}

func LoadHistory() (*History, error) {
	entries, err := LoadAll()
    if err != nil {
        return nil, err
    }
    h := NewHistory(plan.HistoryPath())
    h.Entries = entries
    return h, nil
}
