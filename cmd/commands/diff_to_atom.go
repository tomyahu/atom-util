package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/urfave/cli/v3"
)

var insertBlockTemplateText = `
LINE=$(find-line find "{{.SearchPattern}}" cmd/app.go)
find-line insert -b "{{.BlockFile}}" "{{.TargetFile}}" $LINE
`

type insertBlockData struct {
	SearchPattern string
	BlockFile     string
	TargetFile    string
}

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
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "block_pattern_anything",
			Usage:   "When creating the script.sh file, do regex matching on anything, as opposed to only when the next line of the block is a comment",
			Value:   false,
			Aliases: []string{"m"},
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		diff_file := cmd.StringArg("diff_file")

		data, err := os.ReadFile(diff_file)
		if err != nil {
			return fmt.Errorf("could not read diff file '%s': %w", diff_file, err)
		}

		regex := regexp.MustCompile(`\nindex [\d\w]+\.\.[\d\w]+(?: \d+)?\n--- (.*)\n\+\+\+ b/(.*)\n@@ -(\d+),(\d+) \+(\d+),(\d+) @@[\n ]((:?.*\n)+?)diff`)

		atom_path := fmt.Sprintf("./.atoms/%s", cmd.StringArg("atom_name"))
		err = os.MkdirAll(atom_path, 0755)
		if err != nil {
			return fmt.Errorf("could not create atom directory '%s': %w", atom_path, err)
		}

		err = os.MkdirAll(atom_path+"/tmp_folder", 0755)
		if err != nil {
			return fmt.Errorf("could not create tmp_folder: %w", err)
		}

		tmp_folder_script_file, err := os.OpenFile(atom_path+"/tmp_folder/script.sh", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("could not create tmp_folder script file: %w", err)
		}
		defer tmp_folder_script_file.Close()

		err = createCopierYml(atom_path)
		if err != nil {
			return fmt.Errorf("could not create copier yaml file: %w", err)
		}

		matches := regex.FindAllStringSubmatch(string(data)+"diff", 6)
		for match_index, match := range matches {
			groups_expected := 8
			if len(match) < groups_expected {
				return fmt.Errorf("unexpected number of groups captured when evaluating the regex, expected %d groups and got %d: %s", groups_expected, len(match), match)
			}

			body := match[7]
			target_file := match[2]
			is_new_file := match[1] == "/dev/null"

			if is_new_file {
				file_name := atom_path + "/" + target_file
				err = saveNewFileInAtom(body, file_name)
				if err != nil {
					return err
				}

				continue
			} else {
				added_blocks_regex := regexp.MustCompile(`(?:[\n^]\+.*)+(?:\n.*)`)
				for sm_index, sub_match := range added_blocks_regex.FindAllString(body, -1) {
					file_name := fmt.Sprintf("%s/tmp_folder/%d_%d_%s", atom_path, match_index, sm_index, filepath.Base(target_file))
					err = saveModifiedBlockInAtom(file_name, target_file, sub_match, tmp_folder_script_file, !cmd.Bool("block_pattern_anything"))
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	},
}

// Saves a created file in the diff in the atom
func saveNewFileInAtom(body string, file_name string) error {
	body = strings.ReplaceAll(body, "\n+", "\n")
	plus_regex := regexp.MustCompile(`^\+`)
	minus_regex := regexp.MustCompile(`-.*\n`)

	body = plus_regex.ReplaceAllLiteralString(body, "")
	body = minus_regex.ReplaceAllLiteralString(body, "")

	err := os.MkdirAll(filepath.Dir(file_name), 0755)
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

	return nil
}

// Saves a modified block in an atom and creates a template
func saveModifiedBlockInAtom(file_name string, target_file string, sub_match string, tmp_folder_script_file *os.File, match_only_comments bool) error {
	err := os.MkdirAll(filepath.Dir(file_name), 0755)
	if err != nil {
		return fmt.Errorf("could not create output file directory for '%s': %w", file_name, err)
	}

	find_pattern_regex := regexp.MustCompile(`\n.*$`)
	pattern := find_pattern_regex.FindString(sub_match)
	pattern = strings.Trim(pattern, "\n \t")
	if pattern == "" || (match_only_comments && !strings.HasPrefix(pattern, "//") && !strings.HasPrefix(pattern, "#")) {
		pattern = "1"
	}

	sub_match = strings.ReplaceAll(sub_match, "\n+", "\n")
	sub_match = strings.Replace(sub_match, "\n", "", 1)
	sub_match = find_pattern_regex.ReplaceAllString(sub_match, "")

	file, err := os.OpenFile(file_name, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("could not create output block file '%s': %w", file_name, err)
	}
	defer file.Close()

	_, err = file.WriteString(sub_match)
	if err != nil {
		return fmt.Errorf("could not write to output file '%s': %w", file_name, err)
	}

	err = getInsertBlockTemplate().Execute(tmp_folder_script_file, insertBlockData{
		SearchPattern: pattern,
		BlockFile:     file_name,
		TargetFile:    target_file,
	})
	if err != nil {
		return fmt.Errorf("could not insert block in script file: %w", err)
	}

	return nil
}

func getInsertBlockTemplate() *template.Template {
	return template.Must(template.New("insert_block_template").Parse(insertBlockTemplateText))
}

func createCopierYml(atom_path string) error {
	copier_contents := `

_tasks:
  - ./tmp_folder/script.sh
  - rm -rf tmp_folder
`
	err := os.WriteFile(atom_path+"/copier.yml", []byte(copier_contents), 0755)
	return err
}
