package notes

import (
	"time"

	"github.com/remote-remote/flow/internal/config"
)

// OpenDaily resolves today's daily note, creating it from template if missing.
func OpenDaily(cfg *config.Config, noOpen bool) error {
	path, err := ensureDaily(cfg, time.Now())
	if err != nil {
		return err
	}
	return deliver(path, noOpen)
}

// AppendStandup writes standup content under the ## Standup section of today's daily note.
func AppendStandup(cfg *config.Config, content string) error {
	return appendToDaily(cfg, "## Standup", content, "")
}
