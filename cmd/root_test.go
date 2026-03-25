package cmd

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetUpLogs(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expected    logrus.Level
		expectError bool
	}{
		{name: "debug", level: "debug", expected: logrus.DebugLevel},
		{name: "info", level: "info", expected: logrus.InfoLevel},
		{name: "warn", level: "warn", expected: logrus.WarnLevel},
		{name: "warning", level: "warning", expected: logrus.WarnLevel},
		{name: "error", level: "error", expected: logrus.ErrorLevel},
		{name: "fatal", level: "fatal", expected: logrus.FatalLevel},
		{name: "panic", level: "panic", expected: logrus.PanicLevel},
		{name: "invalid level", level: "invalid", expectError: true},
		{name: "empty string", level: "", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := setUpLogs(&buf, tt.level)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, logrus.GetLevel())
		})
	}
}
