# Changelog

## v1.0.0 (2026-06-28)

Initial release.

- Single binary with subcommands: `start`, `tui`, `config`, `stats`, `version`
- CLI timer with countdown, context-based cancellation, desktop notifications
- TUI timer powered by Bubble Tea and Lipgloss: plan selector, progress bar, pause, skip
- Three built-in presets: classic (4×25min), short (3×15min), long (3×50min)
- Demo preset for recording (`demo.json`)
- Custom JSON plans with repeatable sections and phases
- Plan management: list, add, delete, show, set default
- Persistent configuration via `config.json` in `~/.temgo/`
- Session statistics: today, week, all time; JSON and CSV export
- History: append-only JSONL journal with date-range filtering (`LoadRange`)
- Data stored in `~/.temgo/`: plans, history, config
- Built-in presets via `//go:embed` – self-contained binary
- Custom `Duration` type with JSON marshaling (`"25m"` ↔ `time.Duration`)
- Fluent `Builder` API and `PlanIterator`
- GitHub Actions CI: build, vet, test with race detector
- Demo GIF
- MIT License
