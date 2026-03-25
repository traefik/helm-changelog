package helm

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/traefik/helm-changelog/pkg/git"
)

// GitClient defines the git operations needed to build changelogs.
type GitClient interface {
	GetFileContent(ctx context.Context, hash, filePath string) (string, error)
	GetDiffBetweenCommits(ctx context.Context, start, end, diffPath string) (string, error)
}

// CreateHelmReleases builds a list of releases from git commit history.
func CreateHelmReleases(
	ctx context.Context,
	log *zerolog.Logger,
	chartFile, chartDir string,
	g GitClient,
	commits []git.Commit,
) []*Release {
	var res []*Release

	currentRelease := ""

	var releaseCommits []git.Commit

	log.Info().Msgf(" - Found commits for chart: %d\n", len(commits))

	for _, l := range commits {
		releaseCommits = append(releaseCommits, l)

		chartContent, err := g.GetFileContent(ctx, l.Hash, chartFile)
		if err != nil {
			log.Info().Msgf("Chart.yaml not found in: %s\n", l.Hash)

			continue
		}

		chart, err := GetChart(strings.NewReader(chartContent))
		if err != nil {
			log.Warn().Msgf("Ignoring Chart.yaml file that cannot be parsed: %s", err)

			continue
		}

		if chart.Version != currentRelease {
			log.Info().Msgf(" - Found version: %s\n", chart.Version)

			r := &Release{
				ReleaseDate: l.Author.Date,
				Chart:       chart,
				Commits:     releaseCommits,
			}
			res = append(res, r)
			currentRelease = chart.Version
			releaseCommits = nil
		}
	}

	res = appendUnreleasedCommits(ctx, log, res, releaseCommits, g, chartFile)

	createValueDiffs(ctx, res, g, chartFile, chartDir)

	return res
}

func appendUnreleasedCommits(
	ctx context.Context,
	log *zerolog.Logger,
	res []*Release,
	releaseCommits []git.Commit,
	g GitClient,
	chartFile string,
) []*Release {
	if len(releaseCommits) == 0 {
		return res
	}

	chartContent, err := g.GetFileContent(ctx, "HEAD", chartFile)
	if err != nil {
		return res
	}

	chart, err := GetChart(strings.NewReader(chartContent))
	if err != nil {
		log.Warn().Msgf("Ignoring Chart.yaml file that cannot be parsed: %s", err)

		return res
	}

	chart.Version = "Next Release"

	return append(res, &Release{
		ReleaseDate: nil,
		Chart:       chart,
		Commits:     releaseCommits,
	})
}

func createValueDiffs(
	ctx context.Context,
	res []*Release,
	g GitClient,
	chartFile, chartDir string,
) {
	fullValuesFile := filepath.Join(filepath.Dir(chartFile), "values.yaml")
	relativeValuesFile := filepath.Join(chartDir, "values.yaml")

	for v, release := range res {
		var diff string

		if v > 0 {
			lastRelease := res[v-1]
			lastCommit := lastRelease.Commits[len(lastRelease.Commits)-1].Hash
			currentCommit := release.Commits[len(release.Commits)-1].Hash
			diff, _ = g.GetDiffBetweenCommits(ctx, lastCommit, currentCommit, relativeValuesFile)
		} else {
			diff, _ = g.GetFileContent(ctx, release.Commits[0].Hash, fullValuesFile)
		}

		release.ValueDiff = diff
	}
}
