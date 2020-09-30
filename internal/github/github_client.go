package github

import (
	"context"
	"os"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

// NewGitHubClient creates a github client
func NewGitHubClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GHORG_GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client
}
