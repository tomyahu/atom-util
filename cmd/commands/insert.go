package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

var InsertCommand = &cli.Command{
	Name:               "insert",
	CustomHelpTemplate: SubCommandHelpTemplate,
	Usage:              "Inserts the contents of a source file at a specific position in a target file",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      "source_file",
			UsageText: "The source file to insert in the target file",
		},
		&cli.StringArg{
			Name:      "target_file",
			UsageText: "The target file to insert text in",
		},
		&cli.IntArg{
			Name:      "line_number",
			UsageText: "The line after the thing we want to insert",
		},
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "insert_before_line",
			Usage:   "When included, the tool inserts the contents of the source file before the line specified instead of after",
			Value:   false,
			Aliases: []string{"b"},
		},
		&cli.BoolFlag{
			Name:    "source_as_string",
			Usage:   "When included, the first argument is interpreted as a raw string instead of a file path, and inserted accordingly",
			Value:   false,
			Aliases: []string{"s"},
		},
		&cli.BoolFlag{
			Name:    "append",
			Usage:   "When included, the text is inserted from the end of the selected line. Can be used with the 'at' flag to insert the text in a specific character position",
			Value:   false,
			Aliases: []string{"a"},
		},
		&cli.IntFlag{
			Name:  "at",
			Usage: "To be used with the 'append' flag. Inserts the source text in the same line at the specified position",
			Value: 0,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		sourceFile := cmd.StringArg("source_file")
		targetFile := cmd.StringArg("target_file")
		line := cmd.IntArg("line_number")

		dataToInsert := sourceFile
		if !cmd.Bool("source_as_string") {
			data, err := os.ReadFile(sourceFile)
			if err != nil {
				return fmt.Errorf("could not open source file '%s': %w", sourceFile, err)
			}
			dataToInsert = string(data)
		}

		targetOriginalData, err := os.ReadFile(targetFile)
		if err != nil {
			return fmt.Errorf("could not open target file '%s': %w", targetFile, err)
		}
		splittedTargetOriginalData := strings.Split(string(targetOriginalData), "\n")

		if line < 1 {
			line = max(len(splittedTargetOriginalData) + line)
		}

		result := ""
		for i, lineContent := range splittedTargetOriginalData {
			index := i + 1
			if cmd.Bool("insert_before_line") {
				if index == line {
					result += dataToInsert + "\n"
				}
				result += lineContent + "\n"

			} else {
				result += lineContent + "\n"
				if index == line {
					result += dataToInsert + "\n"
				}
			}
		}

		err = os.WriteFile(targetFile, []byte(result), 0644)
		if err != nil {
			return fmt.Errorf("failed to override target file '%s': %w", targetFile, err)
		}

		return nil
	},
}
