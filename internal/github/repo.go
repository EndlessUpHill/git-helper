package github

import (
	"context"
	"errors"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
)

var (
	ErrTokenNotFound    = errors.New("github token not found")
	ErrInvalidRepoName  = errors.New("invalid repository name format")
	ErrRepoExists      = errors.New("repository already exists")
	ErrUnauthorized    = errors.New("unauthorized: check your GitHub token")
)

type RepoConfig struct {
	Private     bool
	Description string
	Topics      []string
	HasIssues   bool
	HasWiki     bool
}

type Client struct {
	client *github.Client
}

func NewClient(token string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		client: github.NewClient(tc),
	}
}

func (c *Client) CreateRepository(ctx context.Context, name, owner string, isOrg bool, config RepoConfig) error {
	repo := &github.Repository{
		Name:        github.String(name),
		Private:     github.Bool(config.Private),
		Description: github.String(config.Description),
		HasIssues:   github.Bool(config.HasIssues),
		HasWiki:     github.Bool(config.HasWiki),
	}

	var err error
	if isOrg {
		_, _, err = c.client.Repositories.Create(ctx, owner, repo)
	} else {
		_, _, err = c.client.Repositories.Create(ctx, "", repo)
	}

	if err != nil {
		if _, ok := err.(*github.ErrorResponse); ok {
			switch err.(*github.ErrorResponse).Response.StatusCode {
			case 401:
				return ErrUnauthorized
			case 422:
				return ErrRepoExists
			}
		}
		return err
	}

	if len(config.Topics) > 0 {
		_, _, err = c.client.Repositories.ReplaceAllTopics(ctx, owner, name, config.Topics)
	}

	return err
} 