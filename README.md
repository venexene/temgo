# temgo

Таймер для фокусированной работы в терминале. Два режима: CLI и TUI.

## Установка

```
git clone https://github.com/venexene/temgo
cd temgo
go build -o temgo ./cmd/cli
go build -o temgo-tui ./cmd/tui
```

## CLI

```
temgo                    # classic: 4×25мин работы, 3 цикла
temgo -P short           # 3×15мин, 2 цикла
temgo -P long            # 3×50мин, 2 цикла
```

Прерывается по Ctrl+C, история пишется в `.temgo/history.jsonl`.

## TUI

```
temgo-tui
```

Управление: `space` — пауза, `s` — пропустить фазу, `q` — выход.

## Структура

```
cmd/cli          точка входа CLI
cmd/tui          точка входа TUI (Bubble Tea + Lipgloss)
internal/
  plan           модель плана: фазы, секции, builder, итератор
  timer          логика таймера, контекстная остановка
  tui            Bubble Tea Model, рендер, стили
  history        журнал сессий в JSONL
  config         парсинг флагов и пресетов
```



