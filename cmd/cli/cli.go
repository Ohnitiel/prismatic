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
	"ohnitiel/prismatic/internal/locale"

	"github.com/urfave/cli-altsrc/v3"
	toml "github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

var outputFormats = []string{"xlsx", "json", "csv"}

func validateOutputFormat(format string, l *locale.Locale) error {
	if !slices.Contains(outputFormats, strings.ToLower(format)) {
		return fmt.Errorf(l.Errors.OutputFormatNotImpl, format)
	}
	return nil
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

	environments := []string{"production", "replica", "staging"}

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
				Action: func(ctx context.Context, c *cli.Command, s string) error {
					if !slices.Contains(environments, strings.ToLower(s)) {
						return fmt.Errorf(l.Errors.InvalidEnvironment)
					}
					return nil
				},
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
				Usage:     l.CLI.Commands.Export,
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
					query := c.Args().Get(0)
					savePath := c.Args().Get(1)
					if outputFormat == "" {
						outputFormat = filepath.Ext(savePath)
						if outputFormat == "." {
							return fmt.Errorf(l.Errors.OutputFormatEmpty)
						}
						outputFormat = outputFormat[1:]
					} else {
						err := validateOutputFormat(outputFormat, l)
						if err != nil {
							return err
						}
					}
					fmt.Println(query)
					return nil
				},
			},
			{
				Name:      "run",
				Usage:     l.CLI.Commands.Run,
				ArgsUsage: l.CLI.Args.Run,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "commit",
						Usage:       l.CLI.Flags.Commit,
						Destination: &commit,
					},
				},
				// Action: nil,
			},
			{
				Name:  "check",
				Usage: l.CLI.Commands.Check,
				// Action: nil,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
