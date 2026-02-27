package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"ohnitiel/prismatic/internal/config"
	"ohnitiel/prismatic/internal/db"
	"ohnitiel/prismatic/internal/export"
	"ohnitiel/prismatic/internal/locale"

	"github.com/urfave/cli-altsrc/v3"
	toml "github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

const (
	ExitCodeSuccess = 0
	ExitCodeFullFailure = 101
	ExitCodePartialFailure = 102
)

var outputFormats = []string{"xlsx", "json", "csv"}

func validateOutputFormat(format string, l *locale.Locale) error {
	if !slices.Contains(outputFormats, strings.ToLower(format)) {
		return fmt.Errorf(l.Errors.OutputFormatNotImpl, format)
	}
	return nil
}

func startQueryingProcess(
	ctx context.Context, cfg *config.Config, query string,
	environment string, noCache bool, commit bool, command string,
) (map[string]*db.ResultSet, map[string]error) {
	manager := db.NewDatabaseManager()
	manager.LoadConnections(ctx, cfg, environment)

	executor := db.NewExecutor(manager)
	return executor.ParallelExecution(
		ctx, cfg.MaxWorkers, query,
		!noCache, commit, cfg, command,
	)
}

func Prismatic(cfg *config.Config) {
	var environment string
	var config string
	var outputFormat string
	var noSingleSheet bool
	var noSingleFile bool
	var connections []string
	var commit bool
	var noCache bool

	l, err := locale.Load(cfg.Locale)
	if err != nil {
		log.Fatal(err)
	}

	cmd := &cli.Command{
		Name:        "prismatic",
		Description: l.CLI.Description,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "./config/config.toml",
				Usage: l.CLI.Flags.Config,
			},
			&cli.StringFlag{
				Name:        "environment",
				Aliases:     []string{"e"},
				Value:       "replica",
				Usage:       l.CLI.Flags.Environment,
				Destination: &environment,
				Sources:     cli.NewValueSourceChain(toml.TOML("", altsrc.StringSourcer("path"))),
			},
			&cli.StringSliceFlag{
				Name:    "connections",
				Aliases: []string{"c"},
				Usage:   l.CLI.Flags.Connections,
				Sources: cli.NewValueSourceChain(
					toml.TOML("", altsrc.NewStringPtrSourcer(&config))),
				Destination: &connections,
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "export",
				ArgsUsage: l.CLI.Args.Export,
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "query",
					},
					&cli.StringArg{
						Name: "output",
					},
				},
				Usage: l.CLI.Commands.Export,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "output-format",
						Usage: l.CLI.Flags.OutputFormat,
						Action: func(ctx context.Context, c *cli.Command, s string) error {
							return validateOutputFormat(s, l)
						},
						Destination: &outputFormat,
					},
					&cli.BoolFlag{
						Name:        "no-cache",
						Usage:       l.CLI.Flags.NoCache,
						Destination: &noCache,
					},
				},
				MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{{
					Flags: [][]cli.Flag{
						{
							&cli.BoolFlag{
								Name:        "no-single-sheet",
								Usage:       l.CLI.Flags.NoSingleSheet,
								Value:       false,
								Destination: &noSingleSheet,
							},
						},
						{
							&cli.BoolFlag{
								Name:        "no-single-file",
								Usage:       l.CLI.Flags.NoSingleFile,
								Value:       false,
								Destination: &noSingleFile,
							},
						},
					},
				}},
				Action: func(ctx context.Context, c *cli.Command) error {
					query := c.StringArg("query")
					output := c.StringArg("output")

					if outputFormat == "" {
						outputFormat = filepath.Ext(output)
						if outputFormat == "." {
							return fmt.Errorf("%s", l.Errors.OutputFormatEmpty)
						}
						outputFormat = outputFormat[1:]
					} else {
						err := validateOutputFormat(outputFormat, l)
						if err != nil {
							return err
						}
					}

					data, _ := startQueryingProcess(ctx, cfg, query, environment, noCache, commit, c.Name)
					if len(data) == 0 {
						return fmt.Errorf("%s", l.Errors.NoDataReturned)
					}

					excelOptions := export.NewExcelOptions(
						!noSingleFile, !noSingleSheet, cfg.ConnectionColumnName,
					)

					err = export.Excel(ctx, data, output, excelOptions)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "run",
				Usage: l.CLI.Commands.Run,
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "query",
					},
				},
				ArgsUsage: l.CLI.Args.Run,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "commit",
						Usage:       l.CLI.Flags.Commit,
						Destination: &commit,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					query := c.StringArg("query")

					success, failures :=startQueryingProcess(ctx, cfg, query, environment, noCache, commit, c.Name)

					if len(failures) > 0 && len(success) == 0 {
						return cli.Exit(locale.L.ExitMessages.FullFail, ExitCodeFullFailure)
					} else if len(failures) > 0 {
						return cli.Exit(locale.L.ExitMessages.PartialFail, ExitCodePartialFailure)
					} else {
						return cli.Exit(locale.L.ExitMessages.Success, ExitCodeSuccess)
					}
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
