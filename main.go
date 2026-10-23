package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/cli/go-gh/v2/pkg/repository"

	"github.com/rykala/eagher/internal/github"
	"github.com/rykala/eagher/internal/pins"
	"github.com/rykala/eagher/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "eagher:", err)
		os.Exit(1)
	}
}

func run() error {
	repo, err := repository.Current()
	if err != nil {
		return fmt.Errorf("not inside a GitHub repository: %w", err)
	}
	client, err := github.NewClient(repo)
	if err != nil {
		return err
	}
	store, err := pins.Open(client.Repo())
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(ui.New(client, store)).Run()
	return err
}
