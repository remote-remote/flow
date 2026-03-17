package linear

import (
	"context"
	"time"
)

type IssueState struct {
	Name     string `json:"name"`
	Category string `json:"type"`
}

type Issue struct {
	ID         string     `json:"id"`
	Identifier string     `json:"identifier"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	State      IssueState `json:"state"`
}

func (i Issue) FilterValue() string { return i.Identifier + " " + i.Title }

const assignedIssuesQuery = `
query {
  viewer {
    assignedIssues(
      filter: { state: { type: { in: ["started", "unstarted", "backlog"] } } }
      first: 50
      orderBy: updatedAt
    ) {
      nodes {
        id
        identifier
        title
        url
        state { name type }
      }
    }
  }
}
`

func (c *Client) AssignedIssues(ctx context.Context) ([]Issue, error) {
	var resp struct {
		Viewer struct {
			AssignedIssues struct {
				Nodes []Issue `json:"nodes"`
			} `json:"assignedIssues"`
		} `json:"viewer"`
	}
	if err := c.do(ctx, assignedIssuesQuery, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Viewer.AssignedIssues.Nodes, nil
}

const issueByIdentifierQuery = `
query($id: String!) {
  issueVcsBranchSearch(branchName: $id) {
    id
    identifier
    title
    url
    state { name type }
  }
}
`

// IssueByIdentifier fetches a single issue by its identifier (e.g. "ENG-123").
// Linear doesn't have a direct identifier lookup, so we use branch search which matches identifiers.
func (c *Client) IssueByIdentifier(ctx context.Context, identifier string) (*Issue, error) {
	// Use the issues filter with identifier instead
	const q = `
query($filter: IssueFilter!) {
  issues(filter: $filter, first: 1) {
    nodes {
      id
      identifier
      title
      url
      state { name type }
    }
  }
}
`
	vars := map[string]any{
		"filter": map[string]any{
			"identifier": map[string]any{"eq": identifier},
		},
	}
	var resp struct {
		Issues struct {
			Nodes []Issue `json:"nodes"`
		} `json:"issues"`
	}
	if err := c.do(ctx, q, vars, &resp); err != nil {
		return nil, err
	}
	if len(resp.Issues.Nodes) == 0 {
		return nil, nil
	}
	return &resp.Issues.Nodes[0], nil
}

const issuesChangedSinceQuery = `
query($since: DateTime!) {
  viewer {
    assignedIssues(
      filter: { updatedAt: { gte: $since } }
      first: 50
      orderBy: updatedAt
    ) {
      nodes {
        id
        identifier
        title
        url
        state { name type }
      }
    }
  }
}
`

func (c *Client) IssuesChangedSince(ctx context.Context, since time.Time) ([]Issue, error) {
	vars := map[string]any{
		"since": since.Format(time.RFC3339),
	}
	var resp struct {
		Viewer struct {
			AssignedIssues struct {
				Nodes []Issue `json:"nodes"`
			} `json:"assignedIssues"`
		} `json:"viewer"`
	}
	if err := c.do(ctx, issuesChangedSinceQuery, vars, &resp); err != nil {
		return nil, err
	}
	return resp.Viewer.AssignedIssues.Nodes, nil
}
