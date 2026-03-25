// Package cmd provides the CLI commands for helm-changelog.
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/traefik/helm-changelog/pkg/git"
	"github.com/traefik/helm-changelog/pkg/helm"
	"github.com/traefik/helm-changelog/pkg/output"
)

var changelogFilename string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "helm-changelog",
	Short: "Create changelogs for Helm Charts, based on git history",
	Run: func(_ *cobra.Command, _ []string) {
		ctx := context.Background()
		log := zerolog.Ctx(ctx)

		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get working directory")
		}

		g := &git.Git{Log: log}

		gitBaseDir, err := g.FindGitRepositoryRoot(ctx)
		if err != nil {
			log.Fatal().Msg(
				"Could not determine git root directory. helm-changelog depends largely on git history.",
			)
		}

		fileList, err := helm.FindCharts(currentDir)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to find charts")
		}

		for _, chartFileFullPath := range fileList {
			log.Info().Msgf("Handling: %s", chartFileFullPath)

			fullChartDir := filepath.Dir(chartFileFullPath)
			chartFile := strings.TrimPrefix(chartFileFullPath, gitBaseDir+"/")
			relativeChartFile := strings.TrimPrefix(chartFileFullPath, currentDir+"/")
			relativeChartDir := filepath.Dir(relativeChartFile)

			allCommits, err := g.GetAllCommits(ctx, fullChartDir)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get commits")
			}

			releases := helm.CreateHelmReleases(ctx, log, chartFile, relativeChartDir, g, allCommits)

			changeLogFilePath := filepath.Join(fullChartDir, changelogFilename)
			output.Markdown(log, changeLogFilePath, releases)
		}
	},
}

// Execute sets all flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	var v string

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		log, err := setUpLogs(v)
		if err != nil {
			return err
		}

		ctx := log.WithContext(cmd.Context())
		cmd.SetContext(ctx)

		return nil
	}

	rootCmd.PersistentFlags().StringVarP(
		&changelogFilename, "filename", "f", "Changelog.md", "Filename for changelog",
	)
	rootCmd.PersistentFlags().StringVarP(
		&v, "verbosity", "v", zerolog.WarnLevel.String(),
		"Log level (trace, debug, info, warn, error, fatal, panic)",
	)

	cobra.CheckErr(rootCmd.Execute())
}

// setUpLogs configures a zerolog logger and returns it.
func setUpLogs(level string) (zerolog.Logger, error) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("parsing log level: %w", err)
	}

	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		Level(lvl).
		With().
		Timestamp().
		Logger()

	return log, nil
}
