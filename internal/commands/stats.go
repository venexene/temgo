package commands

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
