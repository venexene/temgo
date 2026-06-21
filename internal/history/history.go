package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
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
