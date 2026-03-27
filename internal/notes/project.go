package notes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/remote-remote/flow/internal/config"
)

// ProjectNotePath returns the path for a project note.
func ProjectNotePath(vaultPath, projectName string) string {
	return filepath.Join(vaultPath, "Projects", projectName, projectName+".md")
}

// OpenProject creates a project note if it doesn't exist and opens it in $EDITOR.
func OpenProject(cfg *config.Config, projectName string) error {
	projPath := ProjectNotePath(cfg.VaultPath, projectName)

	if _, err := os.Stat(projPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(projPath), 0o755); err != nil {
			return err
		}
		content := renderProjectTemplate(projectName)
		if err := os.WriteFile(projPath, []byte(content), 0o644); err != nil {
			return err
		}
	}

	return openInEditor(projPath)
}

func renderProjectTemplate(name string) string {
	return fmt.Sprintf(`---
title: "%s"
tags: [project]
---
# %s

## Overview

## Tasks

## Notes
`, name, name)
}
