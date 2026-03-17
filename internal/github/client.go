package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type PR struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	State  string `json:"state"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
}

type Commit struct {
	SHA     string `json:"oid"`
	Message string `json:"messageHeadline"`
}

func gh(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh %v: %s", args, exitErr.Stderr)
		}
		return nil, err
	}
	return out, nil
}

// PRsOpenedOrMerged returns PRs authored by the current user that were updated since the given time.
func PRsOpenedOrMerged(since time.Time) ([]PR, error) {
	date := since.Format("2006-01-02")
	out, err := gh("search", "prs", "--author=@me", "--updated", ">="+date,
		"--json", "title,url,state,author")
	if err != nil {
		return nil, err
	}
	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, err
	}
	return prs, nil
}

// PRsRequestingReview returns PRs where the current user has a pending review request.
func PRsRequestingReview() ([]PR, error) {
	out, err := gh("search", "prs", "--review-requested=@me", "--state=open",
		"--json", "title,url,state,author")
	if err != nil {
		return nil, err
	}
	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, err
	}
	return prs, nil
}

// CommitsPushedSince returns commits by the current user pushed since the given time.
func CommitsPushedSince(since time.Time) ([]Commit, error) {
	date := since.Format("2006-01-02")
	// Use gh to search commits by the authenticated user
	out, err := gh("search", "commits", "--author=@me", "--author-date", ">="+date,
		"--json", "oid,messageHeadline")
	if err != nil {
		return nil, err
	}
	var commits []Commit
	if err := json.Unmarshal(out, &commits); err != nil {
		return nil, err
	}
	return commits, nil
}
