package notes

import (
	"os"
	"os/exec"
	"path/filepath"
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
