// Package cmd provides the CLI commands for helm-changelog.
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/traefik/helm-changelog/pkg/git"
	"github.com/traefik/helm-changelog/pkg/helm"
	"github.com/traefik/helm-changelog/pkg/output"
	"github.com/urfave/cli/v3"
)

// Execute runs the helm-changelog CLI.
func Execute() {
	app := &cli.Command{
		Name:  "helm-changelog",
		Usage: "Create changelogs for Helm Charts, based on git history",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "filename",
				Aliases: []string{"f"},
				Value:   "Changelog.md",
				Usage:   "Filename for changelog",
			},
			&cli.StringFlag{
				Name:    "verbosity",
				Aliases: []string{"v"},
				Value:   zerolog.WarnLevel.String(),
				Usage:   "Log level (trace, debug, info, warn, error, fatal, panic)",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			log, err := setUpLogs(cmd.String("verbosity"))
			if err != nil {
				return ctx, err
			}

			return log.WithContext(ctx), nil
		},
		Action: run,
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	log := zerolog.Ctx(ctx)
	changelogFilename := cmd.String("filename")

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	g := &git.Git{Log: log}

	gitBaseDir, err := g.FindGitRepositoryRoot(ctx)
	if err != nil {
		return fmt.Errorf(
			"could not determine git root directory; helm-changelog depends largely on git history: %w", err,
		)
	}

	fileList, err := helm.FindCharts(currentDir)
	if err != nil {
		return fmt.Errorf("finding charts: %w", err)
	}

	for _, chartFileFullPath := range fileList {
		log.Info().Msgf("Handling: %s", chartFileFullPath)

		fullChartDir := filepath.Dir(chartFileFullPath)
		chartFile := strings.TrimPrefix(chartFileFullPath, gitBaseDir+"/")
		relativeChartFile := strings.TrimPrefix(chartFileFullPath, currentDir+"/")
		relativeChartDir := filepath.Dir(relativeChartFile)

		allCommits, err := g.GetAllCommits(ctx, fullChartDir)
		if err != nil {
			return fmt.Errorf("getting commits: %w", err)
		}

		releases := helm.CreateHelmReleases(ctx, log, chartFile, relativeChartDir, g, allCommits)

		changeLogFilePath := filepath.Join(fullChartDir, changelogFilename)
		output.Markdown(log, changeLogFilePath, releases)
	}

	return nil
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
