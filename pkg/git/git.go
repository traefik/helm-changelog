package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
	"go.yaml.in/yaml/v4"
)

var (
	gitRefPattern = regexp.MustCompile(`^[a-zA-Z0-9_./:^~\-]+$`)

	// ErrInvalidGitRef indicates a git ref contains unexpected characters.
	ErrInvalidGitRef = errors.New("invalid git ref")
)

// Git wraps git CLI operations.
type Git struct {
	Log *zerolog.Logger
}

// FindGitRepositoryRoot returns the absolute path to the root of the git repository.
func (g *Git) FindGitRepositoryRoot(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	g.Log.Debug().Msgf("%s", cmd)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("finding git root: %w", err)
	}

	g.Log.Debug().Msgf("Result: %s", out)

	return strings.TrimSpace(string(out)), nil
}

// GetFileContent returns the content of a file at a specific git revision.
func (g *Git) GetFileContent(ctx context.Context, hash, filePath string) (string, error) {
	err := validateGitRef(hash)
	if err != nil {
		return "", err
	}

	cleanPath := filepath.Clean(filePath)

	cmd := exec.CommandContext(ctx, "git", "cat-file", "-p", hash+":"+cleanPath)
	g.Log.Debug().Msgf("%s", cmd)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting file %s at %s: %w", cleanPath, hash, err)
	}

	g.Log.Debug().Msgf("Result: %s", out)

	return string(out), nil
}

// GetAllCommits returns all git commits affecting the given chart path.
func (g *Git) GetAllCommits(ctx context.Context, chartPath string) ([]Commit, error) {
	cleanPath := filepath.Clean(chartPath)

	cmd := exec.CommandContext(ctx,
		"git",
		"log",
		"--date=iso-strict",
		"--reverse",
		gitformat,
		"--",
		cleanPath,
		":(exclude)"+cleanPath+"/Changelog.md",
	)
	g.Log.Debug().Msgf("%s", cmd)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("getting commits for %s: %w", cleanPath, err)
	}

	if len(out) == 0 {
		return nil, nil
	}

	g.Log.Debug().Msgf("Result: %s", out)

	var commits []Commit

	dec := yaml.NewDecoder(bytes.NewReader(out))

	for {
		t := new(Commit)

		decErr := dec.Decode(&t)

		if t == nil {
			continue
		}

		if errors.Is(decErr, io.EOF) {
			break
		}

		if decErr != nil {
			g.Log.Error().Err(decErr).Msg("failed to decode commit")

			continue
		}

		g.Log.Debug().Msgf("commit: %s %s", t.Hash, t.Subject)

		commits = append(commits, *t)
	}

	return commits, nil
}

// GetDiffBetweenCommits returns the diff of a file between two commits.
func (g *Git) GetDiffBetweenCommits(ctx context.Context, start, end, diffPath string) (string, error) {
	if start == end {
		return "", nil
	}

	err := validateGitRef(start)
	if err != nil {
		return "", err
	}

	err = validateGitRef(end)
	if err != nil {
		return "", err
	}

	cleanPath := filepath.Clean(diffPath)

	cmd := exec.CommandContext(ctx,
		"git",
		"--no-pager",
		"diff",
		start+"..."+end,
		"--",
		cleanPath,
	)
	g.Log.Debug().Msgf("%s", cmd)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting diff between %s and %s: %w", start, end, err)
	}

	g.Log.Debug().Msgf("Result: %s", out)

	return string(out), nil
}

func validateGitRef(ref string) error {
	if !gitRefPattern.MatchString(ref) {
		return fmt.Errorf("%w: %q", ErrInvalidGitRef, ref)
	}

	return nil
}
