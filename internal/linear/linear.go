// Package linear wraps the schpet/linear-cli binary for issue data.
package linear

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type IssueState struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Issue struct {
	ID         string     `json:"id"`
	Identifier string     `json:"identifier"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	BranchName string     `json:"branchName"`
	State      IssueState `json:"state"`
}

func (i Issue) FilterValue() string { return i.Identifier + " " + i.Title }

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (p Project) FilterValue() string { return p.Name }

func resolveLinearCLI() string {
	if p, err := exec.LookPath("linear"); err == nil {
		return p
	}
	home, _ := os.UserHomeDir()
	for _, candidate := range []string{
		filepath.Join(home, ".cargo", "bin", "linear"),
		"/opt/homebrew/bin/linear",
		"/usr/local/bin/linear",
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "linear"
}

var linearCLIPath = resolveLinearCLI()

func linearCLI(args ...string) ([]byte, error) {
	cmd := exec.Command(linearCLIPath, args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("linear %s: %s", strings.Join(args, " "), exitErr.Stderr)
		}
		return nil, err
	}
	return out, nil
}

func graphQL(query string) (json.RawMessage, error) {
	return graphQLWithVars(query, nil)
}

func graphQLWithVars(query string, vars map[string]string) (json.RawMessage, error) {
	args := []string{"api"}
	for k, v := range vars {
		args = append(args, "--variable", k+"="+v)
	}
	args = append(args, query)
	out, err := linearCLI(args...)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("linear API: %s", resp.Errors[0].Message)
	}
	return resp.Data, nil
}

// Projects returns all projects in the workspace.
func Projects() ([]Project, error) {
	out, err := linearCLI("project", "list", "--json")
	if err != nil {
		return nil, err
	}
	var projects []Project
	if err := json.Unmarshal(out, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// ProjectIssues returns all issues for a given project (including unassigned).
func ProjectIssues(projectName string) ([]Issue, error) {
	const query = `query($name: String!) {
		issues(first: 50, filter: { project: { name: { eq: $name } }, state: { type: { in: ["started", "unstarted", "backlog", "triage"] } } }, orderBy: updatedAt) {
			nodes { id identifier title url state { name type } }
		}
	}`
	data, err := graphQLWithVars(query, map[string]string{"name": projectName})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Issues struct {
			Nodes []Issue `json:"nodes"`
		} `json:"issues"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Issues.Nodes, nil
}

// AssignedIssues returns issues assigned to the current user in active states.
func AssignedIssues() ([]Issue, error) {
	const query = `{
		viewer {
			assignedIssues(first: 50, filter: { state: { type: { in: ["started", "unstarted", "backlog"] } } }, orderBy: updatedAt) {
				nodes { id identifier title url state { name type } }
			}
		}
	}`
	data, err := graphQL(query)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Viewer struct {
			AssignedIssues struct {
				Nodes []Issue `json:"nodes"`
			} `json:"assignedIssues"`
		} `json:"viewer"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Viewer.AssignedIssues.Nodes, nil
}

// RecentIssues returns issues assigned to the current user that were recently active
// (in progress, or updated in the last few days), including completed ones.
func RecentIssues() ([]Issue, error) {
	since := time.Now().AddDate(0, 0, -3).Format(time.RFC3339)
	query := fmt.Sprintf(`{
		active: viewer {
			assignedIssues(first: 50, filter: { state: { type: { in: ["started", "unstarted"] } } }, orderBy: updatedAt) {
				nodes { id identifier title url state { name type } }
			}
		}
		recent: viewer {
			assignedIssues(first: 50, filter: { updatedAt: { gte: "%s" } }, orderBy: updatedAt) {
				nodes { id identifier title url state { name type } }
			}
		}
	}`, since)
	data, err := graphQL(query)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Active struct {
			AssignedIssues struct {
				Nodes []Issue `json:"nodes"`
			} `json:"assignedIssues"`
		} `json:"active"`
		Recent struct {
			AssignedIssues struct {
				Nodes []Issue `json:"nodes"`
			} `json:"assignedIssues"`
		} `json:"recent"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var issues []Issue
	for _, list := range [][]Issue{resp.Active.AssignedIssues.Nodes, resp.Recent.AssignedIssues.Nodes} {
		for _, iss := range list {
			if !seen[iss.ID] {
				seen[iss.ID] = true
				issues = append(issues, iss)
			}
		}
	}
	return issues, nil
}

// IssueByIdentifier fetches a single issue by its identifier (e.g. "ENG-123").
func IssueByIdentifier(identifier string) (*Issue, error) {
	out, err := linearCLI("issue", "view", identifier, "--json")
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(out, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// IssuesWorkedSince returns assigned issues that are in progress or were completed since the given time.
func IssuesWorkedSince(since time.Time) ([]Issue, error) {
	ts := since.Format(time.RFC3339)
	// Two queries: currently started issues, and recently completed/canceled issues
	query := fmt.Sprintf(`{
		started: viewer {
			assignedIssues(first: 50, filter: { state: { type: { eq: "started" } }, updatedAt: { gte: "%s" } }, orderBy: updatedAt) {
				nodes { id identifier title url state { name type } }
			}
		}
		completed: viewer {
			assignedIssues(first: 50, filter: { completedAt: { gte: "%s" } }, orderBy: updatedAt) {
				nodes { id identifier title url state { name type } }
			}
		}
	}`, ts, ts)
	data, err := graphQL(query)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Started struct {
			AssignedIssues struct {
				Nodes []Issue `json:"nodes"`
			} `json:"assignedIssues"`
		} `json:"started"`
		Completed struct {
			AssignedIssues struct {
				Nodes []Issue `json:"nodes"`
			} `json:"assignedIssues"`
		} `json:"completed"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	// Merge and deduplicate
	seen := make(map[string]bool)
	var issues []Issue
	for _, list := range [][]Issue{resp.Started.AssignedIssues.Nodes, resp.Completed.AssignedIssues.Nodes} {
		for _, iss := range list {
			if !seen[iss.ID] {
				seen[iss.ID] = true
				issues = append(issues, iss)
			}
		}
	}
	return issues, nil
}

// StartIssue sets the issue to In Progress and assigns to the current user
// without any git operations.
func StartIssue(identifier string) error {
	_, err := linearCLI("issue", "update", identifier, "--state", "In Progress", "--assignee", "self")
	return err
}

// StartIssueWithCheckout uses `linear issue start` to set state, assign,
// and create+checkout the branch in one step.
func StartIssueWithCheckout(identifier string) error {
	_, err := linearCLI("issue", "start", identifier)
	return err
}

// CheckoutBranch checks out an existing git branch for an issue (for resuming work).
func CheckoutBranch(identifier string) error {
	issue, err := IssueByIdentifier(identifier)
	if err != nil {
		return err
	}
	if issue.BranchName == "" {
		return fmt.Errorf("no branch found for %s", identifier)
	}
	cmd := exec.Command("git", "checkout", issue.BranchName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout: %s", out)
	}
	return nil
}
