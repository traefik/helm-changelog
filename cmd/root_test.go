package cmd

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetUpLogs(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expected    zerolog.Level
		expectError bool
	}{
		{name: "trace", level: "trace", expected: zerolog.TraceLevel},
		{name: "debug", level: "debug", expected: zerolog.DebugLevel},
		{name: "info", level: "info", expected: zerolog.InfoLevel},
		{name: "warn", level: "warn", expected: zerolog.WarnLevel},
		{name: "error", level: "error", expected: zerolog.ErrorLevel},
		{name: "fatal", level: "fatal", expected: zerolog.FatalLevel},
		{name: "panic", level: "panic", expected: zerolog.PanicLevel},
		{name: "empty string", level: "", expected: zerolog.NoLevel},
		{name: "invalid level", level: "invalid", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := setUpLogs(tt.level)

			if tt.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, log.GetLevel())
		})
	}
}
