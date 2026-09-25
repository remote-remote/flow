# Flow — Claude Code Notes

## Build & Test

```bash
go build ./...
go test ./...
```

## TUI Gotchas

### Bubbles list.Model needs initial sizing

When creating a `list.Model` inside a delegate sub-model, you must set its size from the parent's dimensions immediately on construction. The default `0, 0` size renders nothing, and delegate sub-models don't reliably receive a `WindowSizeMsg`.

**Pattern:** In every `delegateTo*()` function, call `sub.setSize(m.width, m.height)` (or manually apply `docStyle.GetFrameSize()` offsets) right after creating the sub-model. Don't rely on a future `WindowSizeMsg` to fix it.

### Bubbles textinput.Model needs width and placeholder styling

`textinput.Model` defaults to a narrow width and invisible placeholder text on dark terminals. Always call `SetWidth()` and style the placeholder with a visible color.

**Pattern:** Follow `quick_note.go` — after creating a `textinput.New()`:
```go
ti.SetWidth(60)
styles := ti.Styles()
styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
ti.SetStyles(styles)
```
