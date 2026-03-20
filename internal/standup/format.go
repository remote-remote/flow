package standup

import (
	"strings"
)

// Format renders standup items as a bullet list with URLs.
func Format(data StandupData) string {
	var b strings.Builder

	b.WriteString("- Yesterday\n")
	if len(data.Yesterday) == 0 {
		b.WriteString("    - (none)\n")
	}
	for _, item := range data.Yesterday {
		b.WriteString("    - " + formatItem(item) + "\n")
	}

	b.WriteString("- Today\n")
	if len(data.Today) == 0 {
		b.WriteString("    - (none)\n")
	}
	for _, item := range data.Today {
		b.WriteString("    - " + formatItem(item) + "\n")
	}

	return b.String()
}

func formatItem(item Item) string {
	if item.URL != "" {
		return item.URL
	}
	return item.Text
}
