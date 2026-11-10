package commands

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli/v3"
)

var FindCommand = &cli.Command{
	Name:                          "find",
	Usage:                         "Takes a regex to find the line number of the first match",
	CustomHelpTemplate: SubCommandHelpTemplate,
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
		targetFile := cmd.StringArg("target_file")
		data, err := os.ReadFile(targetFile)
		if err != nil {
			return fmt.Errorf("could not read file '%s': %w", targetFile, err)
		}

		lines := strings.Split(string(data), "\n")

		rawRegex := cmd.StringArg("regex")
		if cmd.Bool("case_insensitive") {
			rawRegex = fmt.Sprintf("(?i)%s", rawRegex)
		}

		regex, err := regexp.Compile(rawRegex)
		if err != nil {
			return fmt.Errorf("could not parse regex '%s': %w", rawRegex, err)
		}

		startLine := cmd.Int("start_line")
		if startLine < 1 {
			startLine = max(len(lines)+startLine, 1)
		}
		endLine := cmd.Int("end_line")
		if endLine < 1 {
			endLine = max(len(lines)+endLine, 1)
		}

		for i, line := range lines {
			index := i + 1
			if index < startLine || index >= endLine {
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
