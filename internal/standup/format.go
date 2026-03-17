package standup

import (
	"fmt"
	"regexp"
	"strings"
)

var identifierRe = regexp.MustCompile(`\[([A-Z]+-\d+)\]`)

// Format renders standup items as markdown with links.
func Format(data StandupData) string {
	var b strings.Builder

	b.WriteString("## Yesterday\n")
	if len(data.Yesterday) == 0 {
		b.WriteString("- (none)\n")
	}
	for _, item := range data.Yesterday {
		b.WriteString("- " + linkify(item) + "\n")
	}

	b.WriteString("\n## Today\n")
	if len(data.Today) == 0 {
		b.WriteString("- (none)\n")
	}
	for _, item := range data.Today {
		b.WriteString("- " + linkify(item) + "\n")
	}

	return b.String()
}

// linkify replaces [ENG-123] with a markdown link if the item has a URL.
func linkify(item Item) string {
	if item.URL != "" {
		if strings.HasPrefix(item.Text, "PR: ") {
			return fmt.Sprintf("[%s](%s)", item.Text, item.URL)
		}
		return identifierRe.ReplaceAllStringFunc(item.Text, func(match string) string {
			id := match[1 : len(match)-1]
			return fmt.Sprintf("[%s](%s)", id, item.URL)
		})
	}
	return item.Text
}
