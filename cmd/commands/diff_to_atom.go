package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v3"
)

var DiffToAtomCommand = &cli.Command{
	Name:               "diff_to_atom",
	Usage:              "takes a diff file and creates an atom",
	CustomHelpTemplate: SubCommandHelpTemplate,
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      "diff_file",
			UsageText: "The file to calculate the diff on",
		},
		&cli.StringArg{
			Name:      "atom_name",
			UsageText: "The name of the atom to create",
		},
	},
	Flags: []cli.Flag{},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		diff_file := cmd.StringArg("diff_file")

		data, err := os.ReadFile(diff_file)
		if err != nil {
			return fmt.Errorf("could not read diff file '%s': %w", diff_file, err)
		}

		regex := regexp.MustCompile(`\nindex [\d\w]+\.\.[\d\w]+(?: \d+)?\n--- (.*)\n\+\+\+ b/(.*)\n@@ -(\d+),(\d+) \+(\d+),(\d+) @@\n((:?.*\n)+?)diff`)

		atom_path := fmt.Sprintf("./.atoms/%s", cmd.StringArg("atom_name"))
		err = os.MkdirAll(atom_path, 0755)
		if err != nil {
			return fmt.Errorf("could not create atom directory '%s': %w", atom_path, err)
		}

		matches := regex.FindAllStringSubmatch(string(data), 6)
		for match_index, match := range matches {
			groups_expected := 8
			if len(match) < groups_expected {
				return fmt.Errorf("unexpected number of groups captured when evaluating the regex, expected %d groups and got %d: %s", groups_expected, len(match), match)
			}

			body := match[7]
			is_new_file := match[1] == "/dev/null"

			if is_new_file {
				file_name := atom_path + "/" + match[2]
				body = strings.ReplaceAll(body, "\n+", "\n")
				plus_regex := regexp.MustCompile(`^\+`)
				minus_regex := regexp.MustCompile(`-.*\n`)

				body = plus_regex.ReplaceAllLiteralString(body, "")
				body = minus_regex.ReplaceAllLiteralString(body, "")

				err = os.MkdirAll(filepath.Dir(file_name), 0755)
				if err != nil {
					return fmt.Errorf("could not create output file directory for '%s': %w", file_name, err)
				}

				file, err := os.OpenFile(file_name, os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					return fmt.Errorf("could not create output file '%s': %w", file_name, err)
				}
				defer file.Close()

				_, err = file.WriteString(body)
				if err != nil {
					return fmt.Errorf("could not write to output file '%s': %w", file_name, err)
				}

				continue
			} else {
				added_blocks_regex := regexp.MustCompile(`(?:\n\+.*)+`)
				for sm_index, sub_match := range added_blocks_regex.FindAllString(body, -1) {
					file_name := fmt.Sprintf("%s/tmp_folder/%d_%d_%s", atom_path, match_index, sm_index, filepath.Base(match[2]))
					err = os.MkdirAll(filepath.Dir(file_name), 0755)
					if err != nil {
						return fmt.Errorf("could not create output file directory for '%s': %w", file_name, err)
					}

					sub_match = strings.ReplaceAll(sub_match, "\n+", "\n")
					sub_match = strings.Replace(sub_match, "\n", "", 1)

					file, err := os.OpenFile(file_name, os.O_WRONLY|os.O_CREATE, 0644)
					if err != nil {
						return fmt.Errorf("could not create output file '%s': %w", file_name, err)
					}
					defer file.Close()

					_, err = file.WriteString(sub_match)
					if err != nil {
						return fmt.Errorf("could not write to output file '%s': %w", file_name, err)
					}

				}

			}

		}

		return nil
	},
}
