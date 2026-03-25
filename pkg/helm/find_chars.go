// Package helm provides utilities for discovering and processing Helm charts.
package helm

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindCharts locates all "Chart.yaml" files in the current directory, and all sub-directories.
func FindCharts(chartSearchDir string) ([]string, error) {
	var fileList []string

	err := filepath.Walk(chartSearchDir, func(path string, _ os.FileInfo, _ error) error {
		fileName := filepath.Base(path)
		if fileName == "Chart.yaml" {
			fileList = append(fileList, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking chart directory %s: %w", chartSearchDir, err)
	}

	return fileList, nil
}
