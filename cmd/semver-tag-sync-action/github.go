package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v82/github"
)

// GitHubClient interface for testing.
type GitHubClient interface {
	GetRef(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error)
	CreateRef(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error)
	UpdateRef(ctx context.Context, owner, repo, ref string, updateRef github.UpdateRef) (*github.Reference, *github.Response, error)
}

// gitHubClientWrapper wraps the go-github client to implement GitHubClient.
type gitHubClientWrapper struct {
	client *github.Client
}

// NewGitHubClient creates a new GitHub client wrapper.
func NewGitHubClient(token, enterpriseURL string) (GitHubClient, error) {
	var client *github.Client
	if enterpriseURL != "" {
		var err error
		client, err = github.NewClient(nil).WithAuthToken(token).WithEnterpriseURLs(enterpriseURL, enterpriseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub Enterprise client: %w", err)
		}
	} else {
		client = github.NewClient(nil).WithAuthToken(token)
	}
	return &gitHubClientWrapper{client: client}, nil
}

func (g *gitHubClientWrapper) GetRef(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
	return g.client.Git.GetRef(ctx, owner, repo, ref)
}

func (g *gitHubClientWrapper) CreateRef(ctx context.Context, owner, repo string, ref github.CreateRef) (*github.Reference, *github.Response, error) {
	return g.client.Git.CreateRef(ctx, owner, repo, ref)
}

func (g *gitHubClientWrapper) UpdateRef(ctx context.Context, owner, repo, ref string, updateRef github.UpdateRef) (*github.Reference, *github.Response, error) {
	return g.client.Git.UpdateRef(ctx, owner, repo, ref, updateRef)
}

// extractTagFromRef extracts the tag name from a git ref.
func extractTagFromRef(ref string) (string, error) {
	if !strings.HasPrefix(ref, "refs/tags/") {
		return "", fmt.Errorf("ref %q is not a tag (expected refs/tags/...)", ref)
	}
	return strings.TrimPrefix(ref, "refs/tags/"), nil
}

// parseRepository parses a repository string in the format "owner/repo".
func parseRepository(repo string) (owner, name string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format %q (expected owner/repo)", repo)
	}
	return parts[0], parts[1], nil
}
