package commands

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/venexene/temgo/internal/history"
	"github.com/venexene/temgo/internal/plan"
)

const statsUsage = `
Usage: temgo stats [flags]

Show session statistics from history.

Flags:
--today       Show today's sessions
--week        Show this week's sessions
--all         Show all sessions (default)
--json        Output in JSON format
--csv         Output in CSV format

Examples:
temgo stats --today
temgo stats --week --json
temgo stats --all --csv > report.csv
`

func RunStats(args []string) error {
	for _, arg := range args {
        if arg == "-h" || arg == "--help" {
            fmt.Print(statsUsage)
			return nil
        }
    }

    fs := flag.NewFlagSet("temgo", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	all := fs.Bool("all", false, "show all sessions stats")
	today := fs.Bool("today", false, "show today's sessions stats")
	week := fs.Bool("week", false, "show this week's sessions stats")
	jsonFlag := fs.Bool("json", false, "save history to json")
	csvFlag := fs.Bool("csv", false, "save history to csv")

	if err := fs.Parse(args); err != nil {
		fmt.Print(statsUsage)
		return err
	}

	count := 0
	if *today { count++ }
	if *week  { count++ }
	if *all   { count++ }
	if count > 1 {
		return fmt.Errorf("--today, --week, and --all are mutually exclusive")
	}
	if count == 0 {
		*all = true
	}

	if *jsonFlag && *csvFlag {
		return fmt.Errorf("--json and --csv are mutually exclusive")
	}

	var entries []history.Entry
	var title string
	var err error
	switch {
	case *today:
		entries, err = history.LoadToday()
		title = "TODAY"
	case *week:
		entries, err = history.LoadWeek()
		title = "THIS WEEK"
	case *all:
		entries, err = history.LoadAll()
		title = "ALL TIME STATS"
	}
	if err != nil {
		return err
	}

	switch {
	case *jsonFlag:
		printJSON(entries)
	case *csvFlag:
		printCSV(entries)
	}
	printString(entries, title)

	return nil
}

func printJSON(entries []history.Entry) {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printCSV(entries []history.Entry) {
	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"type", "start", "duration_sec", "finished"})
	for _, e := range entries {
		w.Write([]string{
			e.Type,
			e.Start.Format(time.RFC3339),
			strconv.Itoa(e.Duration),
			strconv.FormatBool(e.Finished),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write CSV: %v\n", err)
	}
}

func printString(entries []history.Entry, title string) {
	var sb strings.Builder

	typesCount := make(map[string]time.Duration)
	phasesCount := 0
	finishedCount := 0
 	for _, entry := range entries {
		typesCount[entry.Type] += time.Duration(entry.Duration) * time.Second
		phasesCount += 1
		if entry.Finished {
			finishedCount++
		}
	}

	sb.WriteString(fmt.Sprintf("%s:\n", title))
	sb.WriteString(fmt.Sprintf("Sessions: %d\n", phasesCount))
	for phaseType, count := range typesCount {
		sb.WriteString(fmt.Sprintf("%s: %s\n", phaseType, plan.FormatDuration(count)))
	}
	sb.WriteString(fmt.Sprintf("Finished: %d/%d\n", finishedCount, phasesCount))
	fmt.Fprint(os.Stderr, sb.String())
}


