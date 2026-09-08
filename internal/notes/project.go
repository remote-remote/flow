package notes

import (
	"path/filepath"

	"github.com/remote-remote/flow/internal/config"
)

// ProjectNotePath returns the path for a project note.
func ProjectNotePath(vaultPath, projectName string) string {
	return filepath.Join(vaultPath, "Projects", projectName, projectName+".md")
}

// OpenProject creates a project note if it doesn't exist and hands it to the caller.
func OpenProject(cfg *config.Config, projectName string, noOpen bool) error {
	path := ProjectNotePath(cfg.VaultPath, projectName)
	if err := ensureProject(cfg.VaultPath, projectName); err != nil {
		return err
	}
	return deliver(path, noOpen)
}

func ensureProject(vaultPath, projectName string) error {
	return createFromTemplate(
		vaultPath,
		ProjectNotePath(vaultPath, projectName),
		"project",
		map[string]string{"name": projectName},
	)
}
