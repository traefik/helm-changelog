package helm

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/helm-changelog/pkg/git"
)

func newTestLogger() *zerolog.Logger {
	nop := zerolog.Nop()

	return &nop
}

func makeCommit(hash, subject string) git.Commit {
	now := time.Now()

	return git.Commit{
		Hash:    hash,
		Subject: subject,
		Author:  git.Person{Name: "test", Email: "test@test.com", Date: &now},
	}
}

func chartYAML(version string) string {
	return fmt.Sprintf("apiVersion: v2\nname: my-chart\nversion: %s\n", version)
}

var (
	errNotFound = errors.New("not found")
	chartFile   = filepath.Join("charts", "Chart.yaml")
	valuesFile  = filepath.Join("charts", "values.yaml")
)

func TestCreateHelmReleases_SingleRelease(t *testing.T) {
	g := newGitClientMock(t)
	log := newTestLogger()
	ctx := context.Background()

	commits := []git.Commit{
		makeCommit("aaa", "initial work"),
		makeCommit("bbb", "bump version"),
	}

	g.OnGetFileContent("aaa", chartFile).
		TypedReturns("", errNotFound).Once()
	g.OnGetFileContent("bbb", chartFile).
		TypedReturns(chartYAML("1.0.0"), nil).Once()

	g.OnGetFileContent("aaa", valuesFile).
		TypedReturns("key: val", nil).Once()

	releases := CreateHelmReleases(ctx, log, chartFile, "charts", g, commits)

	require.Len(t, releases, 1)
	assert.Equal(t, "1.0.0", releases[0].Chart.Version)
	assert.Len(t, releases[0].Commits, 2)
	assert.Equal(t, "key: val", releases[0].ValueDiff)
}

func TestCreateHelmReleases_MultipleReleases(t *testing.T) {
	g := newGitClientMock(t)
	log := newTestLogger()
	ctx := context.Background()

	commits := []git.Commit{
		makeCommit("aaa", "initial"),
		makeCommit("bbb", "bump to 2.0.0"),
	}

	g.OnGetFileContent("aaa", chartFile).
		TypedReturns(chartYAML("1.0.0"), nil).Once()
	g.OnGetFileContent("bbb", chartFile).
		TypedReturns(chartYAML("2.0.0"), nil).Once()

	g.OnGetFileContent("aaa", valuesFile).
		TypedReturns("key: val1", nil).Once()
	g.OnGetDiffBetweenCommits("aaa", "bbb", valuesFile).
		TypedReturns("- old\n+ new", nil).Once()

	releases := CreateHelmReleases(ctx, log, chartFile, "charts", g, commits)

	require.Len(t, releases, 2)
	assert.Equal(t, "1.0.0", releases[0].Chart.Version)
	assert.Equal(t, "2.0.0", releases[1].Chart.Version)
	assert.Len(t, releases[0].Commits, 1)
	assert.Len(t, releases[1].Commits, 1)
	assert.Equal(t, "key: val1", releases[0].ValueDiff)
	assert.Equal(t, "- old\n+ new", releases[1].ValueDiff)
}

func TestCreateHelmReleases_UnreleasedCommits(t *testing.T) {
	g := newGitClientMock(t)
	log := newTestLogger()
	ctx := context.Background()

	commits := []git.Commit{
		makeCommit("aaa", "released"),
		makeCommit("bbb", "unreleased change"),
	}

	g.OnGetFileContent("aaa", chartFile).
		TypedReturns(chartYAML("1.0.0"), nil).Once()
	g.OnGetFileContent("bbb", chartFile).
		TypedReturns(chartYAML("1.0.0"), nil).Once()
	g.OnGetFileContent("HEAD", chartFile).
		TypedReturns(chartYAML("1.0.0"), nil).Once()

	g.OnGetFileContent("aaa", valuesFile).
		TypedReturns("key: val", nil).Once()
	g.OnGetDiffBetweenCommits("aaa", "bbb", valuesFile).
		TypedReturns("", nil).Once()

	releases := CreateHelmReleases(ctx, log, chartFile, "charts", g, commits)

	require.Len(t, releases, 2)
	assert.Equal(t, "1.0.0", releases[0].Chart.Version)
	assert.Equal(t, "Next Release", releases[1].Chart.Version)
	assert.Nil(t, releases[1].ReleaseDate)
}

func TestCreateHelmReleases_ChartNotFoundInCommit(t *testing.T) {
	g := newGitClientMock(t)
	log := newTestLogger()
	ctx := context.Background()

	commits := []git.Commit{
		makeCommit("aaa", "no chart here"),
		makeCommit("bbb", "has chart"),
	}

	g.OnGetFileContent("aaa", chartFile).
		TypedReturns("", errNotFound).Once()
	g.OnGetFileContent("bbb", chartFile).
		TypedReturns(chartYAML("1.0.0"), nil).Once()

	g.OnGetFileContent("aaa", valuesFile).
		TypedReturns("", errNotFound).Once()

	releases := CreateHelmReleases(ctx, log, chartFile, "charts", g, commits)

	require.Len(t, releases, 1)
	assert.Equal(t, "1.0.0", releases[0].Chart.Version)
	assert.Len(t, releases[0].Commits, 2)
}

func TestCreateHelmReleases_EmptyCommits(t *testing.T) {
	g := newGitClientMock(t)
	log := newTestLogger()
	ctx := context.Background()

	releases := CreateHelmReleases(ctx, log, chartFile, "charts", g, []git.Commit{})

	assert.Empty(t, releases)
}

func TestCreateHelmReleases_InvalidChartYAML(t *testing.T) {
	g := newGitClientMock(t)
	log := newTestLogger()
	ctx := context.Background()

	commits := []git.Commit{
		makeCommit("aaa", "bad chart"),
		makeCommit("bbb", "good chart"),
	}

	g.OnGetFileContent("aaa", chartFile).
		TypedReturns("{{{invalid", nil).Once()
	g.OnGetFileContent("bbb", chartFile).
		TypedReturns(chartYAML("1.0.0"), nil).Once()

	g.OnGetFileContent("aaa", valuesFile).
		TypedReturns("", errNotFound).Once()

	releases := CreateHelmReleases(ctx, log, chartFile, "charts", g, commits)

	require.Len(t, releases, 1)
	assert.Equal(t, "1.0.0", releases[0].Chart.Version)
	assert.Len(t, releases[0].Commits, 2)
}

func TestCreateHelmReleases_ValueDiffs(t *testing.T) {
	g := newGitClientMock(t)
	log := newTestLogger()
	ctx := context.Background()

	commits := []git.Commit{
		makeCommit("aaa", "v1"),
		makeCommit("bbb", "v2"),
	}

	g.OnGetFileContent("aaa", chartFile).
		TypedReturns(chartYAML("1.0.0"), nil).Once()
	g.OnGetFileContent("bbb", chartFile).
		TypedReturns(chartYAML("2.0.0"), nil).Once()

	g.OnGetFileContent("aaa", valuesFile).
		TypedReturns("key: old", nil).Once()
	g.OnGetDiffBetweenCommits("aaa", "bbb", valuesFile).
		TypedReturns("- key: old\n+ key: new", nil).Once()

	releases := CreateHelmReleases(ctx, log, chartFile, "charts", g, commits)

	require.Len(t, releases, 2)
	assert.Equal(t, "key: old", releases[0].ValueDiff)
	assert.Equal(t, "- key: old\n+ key: new", releases[1].ValueDiff)
}
