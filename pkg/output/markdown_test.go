package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/helm-changelog/pkg/git"
	"github.com/traefik/helm-changelog/pkg/helm"
)

func newTestLogger() *zerolog.Logger {
	nop := zerolog.Nop()

	return &nop
}

func date(month, day int) *time.Time {
	t := time.Date(2025, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	return &t
}

func TestMarkdown(t *testing.T) {
	tests := []struct {
		name       string
		releases   []*helm.Release
		latestOnly bool
		golden     string
	}{
		{
			name: "single release",
			releases: []*helm.Release{
				{
					ReleaseDate: date(1, 15),
					Chart: helm.Chart{
						APIVersion: "v2",
						AppVersion: "3.0.0",
						Name:       "my-chart",
						Version:    "1.0.0",
					},
					Commits: []git.Commit{
						{Subject: "feat: add feature A"},
						{Subject: "fix: resolve bug B"},
					},
				},
			},
			golden: "testdata/single_release.md",
		},
		{
			name: "multiple releases",
			releases: []*helm.Release{
				{
					ReleaseDate: date(1, 10),
					Chart: helm.Chart{
						APIVersion: "v2",
						AppVersion: "2.0.0",
						Name:       "my-chart",
						Version:    "1.0.0",
					},
					Commits: []git.Commit{
						{Subject: "feat: initial release"},
					},
				},
				{
					ReleaseDate: date(2, 20),
					Chart: helm.Chart{
						APIVersion: "v2",
						AppVersion: "2.1.0",
						Name:       "my-chart",
						Version:    "2.0.0",
					},
					Commits: []git.Commit{
						{Subject: "feat: new feature"},
						{Subject: "chore: update deps"},
					},
				},
			},
			golden: "testdata/multiple_releases.md",
		},
		{
			name: "deprecated chart",
			releases: []*helm.Release{
				{
					ReleaseDate: date(3, 1),
					Chart: helm.Chart{
						APIVersion: "v2",
						AppVersion: "1.0.0",
						Name:       "old-chart",
						Version:    "1.0.0",
						Deprecated: true,
					},
					Commits: []git.Commit{
						{Subject: "chore: deprecate chart"},
					},
				},
			},
			golden: "testdata/deprecated_chart.md",
		},
		{
			name: "with value diffs",
			releases: []*helm.Release{
				{
					ReleaseDate: date(1, 1),
					Chart: helm.Chart{
						APIVersion: "v2",
						AppVersion: "1.0.0",
						Name:       "my-chart",
						Version:    "1.0.0",
					},
					Commits: []git.Commit{
						{Subject: "feat: initial"},
					},
					ValueDiff: "- old: value\n+ new: value\n",
				},
			},
			golden: "testdata/with_value_diffs.md",
		},
		{
			name: "unreleased commits",
			releases: []*helm.Release{
				{
					ReleaseDate: date(1, 1),
					Chart: helm.Chart{
						APIVersion: "v2",
						AppVersion: "1.0.0",
						Name:       "my-chart",
						Version:    "1.0.0",
					},
					Commits: []git.Commit{
						{Subject: "feat: released feature"},
					},
				},
				{
					ReleaseDate: nil,
					Chart: helm.Chart{
						APIVersion: "v2",
						AppVersion: "1.0.0",
						Name:       "my-chart",
						Version:    "Next Release",
					},
					Commits: []git.Commit{
						{Subject: "feat: work in progress"},
					},
				},
			},
			golden: "testdata/unreleased_commits.md",
		},
		{
			name: "helm v1 api version",
			releases: []*helm.Release{
				{
					ReleaseDate: date(1, 1),
					Chart: helm.Chart{
						APIVersion: "v1",
						AppVersion: "1.0.0",
						Name:       "legacy-chart",
						Version:    "1.0.0",
					},
					Commits: []git.Commit{
						{Subject: "feat: legacy feature"},
					},
				},
			},
			golden: "testdata/helm_v1_api.md",
		},
		{
			name: "no api version",
			releases: []*helm.Release{
				{
					ReleaseDate: date(1, 1),
					Chart: helm.Chart{
						AppVersion: "1.0.0",
						Name:       "ancient-chart",
						Version:    "1.0.0",
					},
					Commits: []git.Commit{
						{Subject: "feat: ancient feature"},
					},
				},
			},
			golden: "testdata/no_api_version.md",
		},
		{
			name: "with kube version",
			releases: []*helm.Release{
				{
					ReleaseDate: date(1, 1),
					Chart: helm.Chart{
						APIVersion:  "v2",
						AppVersion:  "1.0.0",
						KubeVersion: ">=1.20.0-0",
						Name:        "my-chart",
						Version:     "1.0.0",
					},
					Commits: []git.Commit{
						{Subject: "feat: add feature"},
					},
				},
			},
			golden: "testdata/with_kube_version.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			outPath := filepath.Join(dir, "Changelog.md")

			Markdown(newTestLogger(), outPath, tt.releases, tt.latestOnly)

			got, err := os.ReadFile(outPath)
			require.NoError(t, err)

			if os.Getenv("UPDATE_GOLDEN") == "1" {
				require.NoError(t, os.WriteFile(tt.golden, got, 0o644))
			}

			expected, err := os.ReadFile(tt.golden)
			require.NoError(t, err, "Golden file %s not found. Run with UPDATE_GOLDEN=1 to create it.", tt.golden)

			// Normalize line endings for cross-platform compatibility.
			assert.Equal(t, strings.ReplaceAll(string(expected), "\r\n", "\n"), strings.ReplaceAll(string(got), "\r\n", "\n"))
		})
	}
}

func TestMarkdown_LatestOnly(t *testing.T) {
	existingChangelog, err := os.ReadFile("testdata/existing_changelog.md")
	require.NoError(t, err)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "Changelog.md")

	require.NoError(t, os.WriteFile(outPath, existingChangelog, 0o644))

	// Now generate with latestOnly=true, adding v2.0.0.
	releases := []*helm.Release{
		{
			ReleaseDate: date(1, 10),
			Chart: helm.Chart{
				APIVersion: "v2",
				AppVersion: "1.0.0",
				Name:       "my-chart",
				Version:    "1.0.0",
			},
			Commits: []git.Commit{
				{Subject: "feat: initial release"},
			},
		},
		{
			ReleaseDate: date(2, 20),
			Chart: helm.Chart{
				APIVersion: "v2",
				AppVersion: "2.0.0",
				Name:       "my-chart",
				Version:    "2.0.0",
			},
			Commits: []git.Commit{
				{Subject: "feat: new feature"},
			},
		},
	}

	Markdown(newTestLogger(), outPath, releases, true)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)

	result := strings.ReplaceAll(string(got), "\r\n", "\n")

	// v2.0.0 should be present (newly generated).
	assert.Contains(t, result, "## 2.0.0 ")
	// v1.0.0 should still have the manually edited text.
	assert.Contains(t, result, "manually edited")
	// The header should appear only once.
	assert.Equal(t, 1, strings.Count(result, "# Change Log"))
}
