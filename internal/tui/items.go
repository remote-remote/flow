package tui

import "github.com/remote-remote/flow/internal/linear"

type issueItem struct {
	issue linear.Issue
}

func (i issueItem) Title() string       { return i.issue.Identifier + ": " + i.issue.Title }
func (i issueItem) Description() string { return i.issue.State.Name }
func (i issueItem) FilterValue() string { return i.issue.FilterValue() }

type projectItem struct {
	project linear.Project
}

func (i projectItem) Title() string       { return i.project.Name }
func (i projectItem) Description() string { return "" }
func (i projectItem) FilterValue() string { return i.project.FilterValue() }
