package plan

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a time.Duration that marshals to/from human-readable strings like "25m".
type Duration time.Duration

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// String returns the duration formatted as MM:SS or H:MM:SS.
func (d Duration) String() string {
	return FormatDuration(time.Duration(d))
}

// FormatDuration formats a time.Duration as MM:SS or H:MM:SS.
func FormatDuration(t time.Duration) string {
	seconds := int(t.Seconds())
	if seconds >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", seconds/3600, (seconds%3600)/60, seconds%60)
	} else {
		return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
	}
}
