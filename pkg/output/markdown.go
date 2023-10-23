package output

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mloiseleur/helm-changelog/pkg/helm"
)

// Markdown creates a markdown representation of the changelog at the changeLogFilePath path
func Markdown(log *logrus.Logger, changeLogFilePath string, releases []*helm.Release) {
	// reverse commits
	for _, release := range releases {
		release.Commits = reverseCommits(release.Commits)
	}

	// reverse releases
	releases = reverseReleases(releases)

	log.Debugf("Creating changelog file: %s", changeLogFilePath)
	f, err := os.Create(changeLogFilePath)
	if err != nil {
		log.Fatalf("Failed creating changelog file")
	}

	defer f.Close()

	f.WriteString("# Change Log\n\n")

	for _, release := range releases {

		if release.Chart.Deprecated {
			f.WriteString(fmt.Sprintf("## %s (DEPRECATED)\n\n", release.Chart.Version))
		} else {
			f.WriteString(fmt.Sprintf("## %s ", strings.TrimSpace(release.Chart.Version)))
		}

		if release.Chart.AppVersion != "" {
			f.WriteString(badge("AppVersion", release.Chart.AppVersion, "", "success"))
		}

		if release.Chart.KubeVersion != "" {
			f.WriteString(badge("Kubernetes", release.Chart.KubeVersion, "kubernetes", "informational"))
		}

		if release.Chart.APIVersion == "" {
			f.WriteString(badge("Helm", "v2", "helm", "inactive"))
		}

		if release.Chart.APIVersion == "v1" {
			f.WriteString(badge("Helm", "v2", "helm", "inactive"))
			f.WriteString(badge("Helm", "v3", "helm", "informational"))
		}

		if release.Chart.APIVersion == "v2" {
			f.WriteString(badge("Helm", "v3", "helm", "informational"))
		}

		f.WriteString("\n\n")

		if release.ReleaseDate != nil {
			f.WriteString(fmt.Sprintf("**Release date:** %s\n\n", release.ReleaseDate.Format("2006-01-02")))
		}

		sort.Slice(release.Commits, func(i, j int) bool { return release.Commits[i].Subject > release.Commits[j].Subject })
		for _, l := range release.Commits {
			f.WriteString(fmt.Sprintf("* %s\n", l.Subject))
		}

		for _, l := range release.Commits {
			f.WriteString(fmt.Sprintf("* %s\n", strings.TrimSpace(l.Subject)))
		}
		f.WriteString("\n")

		if release.ValueDiff != "" {
			f.WriteString("### Default value changes\n\n")
			f.WriteString("```diff\n")
			f.WriteString(release.ValueDiff)
			f.WriteString("```\n")
		}

		f.WriteString("\n")
	}

	f.WriteString("---\n")
	// TODO Add version number
	f.WriteString(fmt.Sprintf("Autogenerated from Helm Chart and git history using [helm-changelog](https://github.com/mogensen/helm-changelog)\n"))
}

func badge(key, value, icon, style string) string {
	return fmt.Sprintf(" ![%s: %s](https://img.shields.io/static/v1?label=%s&message=%s&color=%s&logo=%s)", key, value, key, url.QueryEscape(value), style, icon)
}
