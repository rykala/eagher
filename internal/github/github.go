package github

import (
	"context"
	"fmt"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
)

const pageSize = 50

const commonFields = `
	number
	title
	url
	body
	updatedAt
	author { login }
	labels(first: 10) { nodes { name } }`

type Item struct {
	Number    int
	Title     string
	URL       string
	Body      string
	UpdatedAt time.Time
	Author    string
	Labels    []string
	IsDraft   bool
}

type Client struct {
	gql  *api.GraphQLClient
	repo repository.Repository
}

func NewClient(repo repository.Repository) (*Client, error) {
	gql, err := api.NewGraphQLClient(api.ClientOptions{Host: repo.Host})
	if err != nil {
		return nil, err
	}
	return &Client{gql: gql, repo: repo}, nil
}

func (c *Client) Repo() string {
	return c.repo.Owner + "/" + c.repo.Name
}

func (c *Client) Issues(ctx context.Context) ([]Item, error) {
	return c.list(ctx, "issues", commonFields)
}

func (c *Client) PullRequests(ctx context.Context) ([]Item, error) {
	return c.list(ctx, "pullRequests", commonFields+"\n\tisDraft")
}

func (c *Client) list(ctx context.Context, connection, fields string) ([]Item, error) {
	query := fmt.Sprintf(`query($owner: String!, $name: String!, $first: Int!) {
	repository(owner: $owner, name: $name) {
		items: %s(first: $first, states: OPEN, orderBy: {field: UPDATED_AT, direction: DESC}) {
			nodes {%s
			}
		}
	}
}`, connection, fields)

	var resp struct {
		Repository struct {
			Items struct {
				Nodes []struct {
					Number    int
					Title     string
					URL       string
					Body      string
					UpdatedAt time.Time
					IsDraft   bool
					Author    *struct{ Login string }
					Labels    struct {
						Nodes []struct{ Name string }
					}
				}
			}
		}
	}
	vars := map[string]any{"owner": c.repo.Owner, "name": c.repo.Name, "first": pageSize}
	if err := c.gql.DoWithContext(ctx, query, vars, &resp); err != nil {
		return nil, err
	}

	items := make([]Item, 0, len(resp.Repository.Items.Nodes))
	for _, n := range resp.Repository.Items.Nodes {
		author := "ghost"
		if n.Author != nil {
			author = n.Author.Login
		}
		labels := make([]string, 0, len(n.Labels.Nodes))
		for _, l := range n.Labels.Nodes {
			labels = append(labels, l.Name)
		}
		items = append(items, Item{
			Number:    n.Number,
			Title:     n.Title,
			URL:       n.URL,
			Body:      n.Body,
			UpdatedAt: n.UpdatedAt,
			Author:    author,
			Labels:    labels,
			IsDraft:   n.IsDraft,
		})
	}
	return items, nil
}
