package commands

import (
	"fmt"
	"os"
	"strings"
)

var SubCommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .Arguments}}

ARGUMENTS:{{range .Arguments}}
   {{.Name}} - {{.UsageText}} {{end}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}
`

func insertTextAtFileLine(text string, filename string, line int, insertBeforeLine bool) error {
	targetOriginalData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not open target file '%s': %w", filename, err)
	}
	splittedTargetOriginalData := strings.Split(string(targetOriginalData), "\n")

	if line < 1 {
		line = max(len(splittedTargetOriginalData) + line)
	}

	result := []string{}
	for i, lineContent := range splittedTargetOriginalData {
		index := i + 1
		if insertBeforeLine {
			if index == line {
				result = append(result, text)
			}
			result = append(result, lineContent)

		} else {
			result = append(result, lineContent)
			if index == line {
				result = append(result, text)
			}
		}
	}

	err = os.WriteFile(filename, []byte(strings.Join(result, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to override target file '%s': %w", filename, err)
	}

	return nil
}

func getLineContent(data string, line int) (string, error) {
	splittedData := strings.Split(data, "\n")

	if line < 1 || line > len(splittedData) {
		return "", fmt.Errorf("attempted to extract line %d from a string with %d lines", line, len(splittedData))
	}

	return splittedData[line-1], nil
}
