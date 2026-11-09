package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli/v3"
)

var MyHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{template "usageTemplate" .}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .Arguments}}

ARGUMENTS:{{range .Arguments}}
   {{.Name}} - {{.UsageText}} {{end}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}
`

func App() *cli.Command {
	commands := &cli.Command{
		CustomRootCommandHelpTemplate: MyHelpTemplate,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "regex",
				UsageText: "The regex to match to return the line",
			},
			&cli.StringArg{
				Name:      "target_file",
				UsageText: "Target file to search the regex on",
			},
		},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "start_line",
				Usage:   "The line to start searching the regex on. 0 and negative numbers are interpreted as starting from the last line of the file",
				Value:   1,
				Aliases: []string{"s"},
			},
			&cli.IntFlag{
				Name:    "end_line",
				Usage:   "The line to finish searching the regex on. 0 and negative numbers are interpreted as starting from the last line of the file",
				Value:   0,
				Aliases: []string{"e"},
			},
			&cli.BoolFlag{
				Name:    "case_insensitive",
				Usage:   "matches the regex in a case insensitive manner",
				Value:   false,
				Aliases: []string{"i"},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			target_file := cmd.StringArg("target_file")
			data, err := os.ReadFile(target_file)
			if err != nil {
				return fmt.Errorf("could not read file '%s': %w", target_file, err)
			}

			lines := strings.Split(string(data), "\n")

			raw_regex := cmd.StringArg("regex")
			if cmd.Bool("case_insensitive") {
				raw_regex = fmt.Sprintf("(?i)%s", raw_regex)
			}

			regex, err := regexp.Compile(raw_regex)
			if err != nil {
				return fmt.Errorf("could not parse regex '%s': %w", raw_regex, err)
			}

			start_line := cmd.Int("start_line")
			if start_line < 1 {
				start_line = len(lines) - start_line
			}
			end_line := cmd.Int("end_line")
			if end_line < 1 {
				end_line = len(lines) - end_line
			}

			for i, line := range lines {
				index := i + 1
				if index < start_line || index >= end_line {
					continue
				}

				if regex.Match([]byte(line)) {
					fmt.Println(index)
					return nil
				}
			}

			return nil
		},
	}

	return commands
}
