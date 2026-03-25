package helm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetChart(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Chart
		expectError bool
	}{
		{
			name: "full chart",
			input: `apiVersion: v2
appVersion: 1.0.0
kubeVersion: ">=1.20.0-0"
name: my-chart
deprecated: false
description: A test chart
version: 1.2.3`,
			expected: Chart{
				APIVersion:  "v2",
				AppVersion:  "1.0.0",
				KubeVersion: ">=1.20.0-0",
				Name:        "my-chart",
				Deprecated:  false,
				Description: "A test chart",
				Version:     "1.2.3",
			},
		},
		{
			name: "minimal chart",
			input: `name: minimal
version: 0.1.0`,
			expected: Chart{
				Name:    "minimal",
				Version: "0.1.0",
			},
		},
		{
			name: "deprecated chart",
			input: `apiVersion: v2
name: old-chart
version: 1.0.0
deprecated: true`,
			expected: Chart{
				APIVersion: "v2",
				Name:       "old-chart",
				Version:    "1.0.0",
				Deprecated: true,
			},
		},
		{
			name: "v1 api version",
			input: `apiVersion: v1
name: legacy
version: 2.0.0`,
			expected: Chart{
				APIVersion: "v1",
				Name:       "legacy",
				Version:    "2.0.0",
			},
		},
		{
			name:        "invalid yaml",
			input:       `{{{invalid`,
			expectError: true,
		},
		{
			name:        "empty input",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetChart(strings.NewReader(tt.input))

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
