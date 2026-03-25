package helm

import (
	"fmt"
	"io"
	"time"

	"github.com/traefik/helm-changelog/pkg/git"
	"gopkg.in/yaml.v3"
)

// Chart contains all info about a chart from the Chart.yaml file.
type Chart struct {
	APIVersion  string `yaml:"apiVersion"`
	AppVersion  string `yaml:"appVersion"`
	KubeVersion string `yaml:"kubeVersion"`
	Name        string `yaml:"name"`
	Deprecated  bool   `yaml:"deprecated"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

// Release represents a release of a Helm Chart including:
// - metadata for the released chart
// - all commits
// - the difference in default values.
type Release struct {
	ReleaseDate *time.Time
	Chart       Chart
	ValueDiff   string
	Commits     []git.Commit
}

// GetChart reads the content of the io.Reader and unmarshals the content into a Chart struct.
func GetChart(chartFile io.Reader) (Chart, error) {
	var chartMeta Chart

	decoder := yaml.NewDecoder(chartFile)

	err := decoder.Decode(&chartMeta)
	if err != nil {
		return Chart{}, fmt.Errorf("decoding chart: %w", err)
	}

	return chartMeta, nil
}
