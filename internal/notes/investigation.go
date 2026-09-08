package notes

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/remote-remote/flow/internal/config"
)

// Investigation is an append-only working note: the agent writes evidence links
// and decisions into it as they land, and distillation reads it afterwards.
type Investigation struct {
	// Title is the note's heading and, for vault-level notes, its filename.
	Title string

	// Project and Task record what the investigation is for, when that is known.
	// Project also scopes a plan note under Projects/{Project}/.
	Project string
	Task    string

	// Scratch is the repo-keyed directory oracles write evidence to, recorded so
	// the note is self-describing once the branch it started on is gone.
	Scratch string

	// Scopes names the oracles fanned out for this investigation. They double as
	// herdr agent names, which is what makes a restarted oracle resumable.
	Scopes []string

	// plan marks a project plan, which lives beside the project note under a
	// fixed name rather than in the vault-level Notes/ directory.
	plan bool
}

// NewInvestigation describes a vault-level investigation note.
func NewInvestigation(title string) Investigation {
	return Investigation{Title: title}
}

// NewProjectPlan describes an investigation into a project that already exists.
func NewProjectPlan(projectName string) Investigation {
	return Investigation{Title: projectName + " plan", Project: projectName, plan: true}
}

// Path returns where the investigation note lives. A title that slugifies to
// nothing has no filename to give, so it is rejected rather than written to
// Notes/.md.
func (inv Investigation) Path(vaultPath string) (string, error) {
	if inv.plan {
		return filepath.Join(vaultPath, "Projects", inv.Project, "plan.md"), nil
	}
	slug := slugify(inv.Title)
	if slug == "" {
		return "", errors.New("title has no letters or digits to build a filename from")
	}
	return filepath.Join(vaultPath, "Notes", slug+".md"), nil
}

// OpenInvestigation scaffolds the note if it is missing, cross-links it into
// today's daily note, and hands it to the caller.
func OpenInvestigation(cfg *config.Config, inv Investigation, noOpen bool) error {
	path, err := inv.Path(cfg.VaultPath)
	if err != nil {
		return err
	}

	if inv.plan {
		if err := ensureProject(cfg.VaultPath, inv.Project); err != nil {
			return err
		}
	}

	if err := createFromTemplate(cfg.VaultPath, path, "investigation", inv.vars()); err != nil {
		return err
	}

	wikilink := vaultWikilink(cfg.VaultPath, path, inv.Title)
	if err := appendToDaily(cfg, "## Notes", "- "+wikilink, wikilink); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not cross-link to daily note: %v\n", err)
	}

	return deliver(path, noOpen)
}

func (inv Investigation) vars() map[string]string {
	quoted := make([]string, len(inv.Scopes))
	for i, s := range inv.Scopes {
		quoted[i] = `"` + s + `"`
	}
	return map[string]string{
		"title":   inv.Title,
		"date":    time.Now().Format("2006-01-02"),
		"scratch": inv.Scratch,
		"scopes":  strings.Join(quoted, ", "),
		"project": inv.Project,
		"task":    inv.Task,
	}
}

// vaultWikilink builds an Obsidian link to a note from its filesystem path.
func vaultWikilink(vaultPath, notePath, label string) string {
	rel, err := filepath.Rel(vaultPath, notePath)
	if err != nil {
		rel = filepath.Base(notePath)
	}
	return fmt.Sprintf("[[%s|%s]]", strings.TrimSuffix(filepath.ToSlash(rel), ".md"), label)
}
