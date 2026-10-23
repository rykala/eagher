package pins

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// Store keeps the pinned issue and pull request numbers of one repository.
// Issues and pull requests share a single number sequence, so one list per
// repository is enough.
type Store struct {
	path    string
	repo    string
	numbers []int
}

func Open(repo string) (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return open(filepath.Join(dir, "eagher", "pins.json"), repo)
}

func open(path, repo string) (*Store, error) {
	all, err := readAll(path)
	if err != nil {
		return nil, err
	}
	return &Store{path: path, repo: repo, numbers: all[repo]}, nil
}

func (s *Store) Has(number int) bool {
	return slices.Contains(s.numbers, number)
}

func (s *Store) Toggle(number int) error {
	if i := slices.Index(s.numbers, number); i >= 0 {
		s.numbers = slices.Delete(s.numbers, i, i+1)
	} else {
		s.numbers = append(s.numbers, number)
	}
	return s.save()
}

// save re-reads the file so pins changed by eagher running in other
// repositories are not overwritten.
func (s *Store) save() error {
	all, err := readAll(s.path)
	if err != nil {
		return err
	}
	if len(s.numbers) == 0 {
		delete(all, s.repo)
	} else {
		all[s.repo] = s.numbers
	}

	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.path, append(data, '\n'), 0o644)
}

func readAll(path string) (map[string][]int, error) {
	all := map[string][]int{}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return all, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &all); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return all, nil
}
