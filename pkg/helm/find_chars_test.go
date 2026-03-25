package helm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindCharts(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		expected int
	}{
		{
			name:     "empty directory",
			setup:    func(_ *testing.T, _ string) {},
			expected: 0,
		},
		{
			name: "single chart at root",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("name: test"), 0o644))
			},
			expected: 1,
		},
		{
			name: "nested charts",
			setup: func(t *testing.T, dir string) {
				t.Helper()

				charts := []string{
					filepath.Join(dir, "chart-a", "Chart.yaml"),
					filepath.Join(dir, "chart-b", "Chart.yaml"),
					filepath.Join(dir, "charts", "sub", "Chart.yaml"),
				}
				for _, c := range charts {
					require.NoError(t, os.MkdirAll(filepath.Dir(c), 0o755))
					require.NoError(t, os.WriteFile(c, []byte("name: test"), 0o644))
				}
			},
			expected: 3,
		},
		{
			name: "ignores non Chart.yaml files",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("name: test"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"), []byte("key: val"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "chart.yaml"), []byte("name: lower"), 0o644))
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			result, err := FindCharts(dir)
			require.NoError(t, err)
			assert.Len(t, result, tt.expected)

			for _, path := range result {
				assert.Equal(t, "Chart.yaml", filepath.Base(path))
			}
		})
	}
}

func TestFindCharts_NonExistentDirectory(t *testing.T) {
	result, _ := FindCharts("/nonexistent/path")
	assert.Empty(t, result)
}
