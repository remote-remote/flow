package notes

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/remote-remote/flow/internal/config"
)

// OpenDaily resolves today's daily note path, creates it from template if missing, and opens it in $EDITOR.
func OpenDaily(cfg *config.Config) error {
	now := time.Now()

	path, err := config.DailyNotePath(cfg.VaultPath, now)
	if err != nil {
		return err
	}

	// Create file from template if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		content := RenderDailyTemplate(now)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AppendStandup writes standup content under the ## Standup section of today's daily note.
func AppendStandup(cfg *config.Config, content string) error {
	now := time.Now()
	dailyPath, err := config.DailyNotePath(cfg.VaultPath, now)
	if err != nil {
		return err
	}

	// Create daily note if it doesn't exist
	if _, err := os.Stat(dailyPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dailyPath), 0o755); err != nil {
			return err
		}
		tmpl := RenderDailyTemplate(now)
		if err := os.WriteFile(dailyPath, []byte(tmpl), 0o644); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(dailyPath)
	if err != nil {
		return err
	}

	note := string(data)
	standupIdx := strings.Index(note, "## Standup")
	if standupIdx == -1 {
		// Append section if missing
		note += "\n## Standup\n" + content + "\n"
	} else {
		// Insert after the ## Standup line
		afterHeading := standupIdx + len("## Standup")
		nextLine := strings.Index(note[afterHeading:], "\n")
		if nextLine == -1 {
			note += "\n" + content + "\n"
		} else {
			insertAt := afterHeading + nextLine + 1
			note = note[:insertAt] + content + "\n" + note[insertAt:]
		}
	}

	return os.WriteFile(dailyPath, []byte(note), 0o644)
}
