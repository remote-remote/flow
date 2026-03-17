// Package linear wraps the linear-cli binary for issue data.
package linear

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type IssueState struct {
	Name string `json:"name"`
}

type Issue struct {
	ID         string     `json:"id"`
	Identifier string     `json:"identifier"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	State      IssueState `json:"state"`
}

func (i Issue) FilterValue() string { return i.Identifier + " " + i.Title }

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (p Project) FilterValue() string { return p.Name }

func linearCLI(args ...string) ([]byte, error) {
	cmd := exec.Command("linear-cli", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("linear-cli %s: %s", strings.Join(args, " "), exitErr.Stderr)
		}
		return nil, err
	}
	return out, nil
}

// Projects returns all projects in the workspace.
func Projects() ([]Project, error) {
	out, err := linearCLI("p", "list", "--output", "json")
	if err != nil {
		return nil, err
	}
	var projects []Project
	if err := json.Unmarshal(out, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// ProjectIssues returns issues for a given project name.
func ProjectIssues(projectName string) ([]Issue, error) {
	out, err := linearCLI("i", "list", "--project", projectName, "--output", "json")
	if err != nil {
		return nil, err
	}
	var issues []Issue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// AssignedIssues returns issues assigned to the current user.
func AssignedIssues() ([]Issue, error) {
	out, err := linearCLI("i", "list", "--output", "json")
	if err != nil {
		return nil, err
	}
	var issues []Issue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// IssueByIdentifier fetches a single issue by its identifier (e.g. "ENG-123").
func IssueByIdentifier(identifier string) (*Issue, error) {
	out, err := linearCLI("i", "get", identifier, "--output", "json")
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(out, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// IssuesChangedSince returns issues updated within the given duration string (e.g. "1d", "3d").
func IssuesChangedSince(since string) ([]Issue, error) {
	out, err := linearCLI("i", "list", "--since", since, "--output", "json")
	if err != nil {
		return nil, err
	}
	var issues []Issue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// StartIssue assigns the issue to the current user and sets it to In Progress.
func StartIssue(identifier string) error {
	_, err := linearCLI("i", "start", identifier)
	return err
}

// StartIssueWithCheckout starts the issue and checks out the git branch.
func StartIssueWithCheckout(identifier string) error {
	_, err := linearCLI("i", "start", identifier, "--checkout")
	return err
}
