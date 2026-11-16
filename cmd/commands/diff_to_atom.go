package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"example.com/atom-util/cmd/commands/diff_to_atom"
	"github.com/google/uuid"

	"github.com/urfave/cli/v3"
)

var insertBlockTemplateText = `
LINE=$(atom-util find "{{.SearchPattern}}" {{.TargetFile}})
atom-util insert -b "{{.BlockFile}}" "{{.TargetFile}}" $LINE
`

type insertBlockData struct {
	SearchPattern string
	BlockFile     string
	TargetFile    string
}

type insertBlockScriptData struct {
	BlockStartLine int
	BlockContent   string
	Filename       string
	TargetFileName string
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
		&cli.StringFlag{
			Name:    "output_folder",
			Usage:   "The output folder where to save the atom",
			Value:   ".atoms",
			Aliases: []string{"o"},
		},
		&cli.BoolFlag{
			Name:    "add_uuid_comments",
			Usage:   "When specified the command will add a comment containing an uuid to the original file on each place a block should be inserted",
			Value:   false,
			Aliases: []string{"c"},
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		diff_file := cmd.StringArg("diff_file")

		data, err := os.ReadFile(diff_file)
		if err != nil {
			return fmt.Errorf("could not read diff file '%s': %w", diff_file, err)
		}

		atom_path := fmt.Sprintf("%s/%s", cmd.String("output_folder"), cmd.StringArg("atom_name"))
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
		defer tmp_folder_script_file.Close() //nolint:errcheck

		err = createCopierYml(atom_path)
		if err != nil {
			return fmt.Errorf("could not create copier yaml file: %w", err)
		}
		parsedGitDiff, err := diff_to_atom.ParseDiffFile(string(data))
		if err != nil {
			return fmt.Errorf("failed to parse git diff: %w", err)
		}

		for file_index, diffFile := range parsedGitDiff.ModifiedFiles {
			if diffFile.IsNewFile {
				file_name := atom_path + "/" + diffFile.TargetFile
				err = saveNewFileInAtom(diffFile.Blocks[0].Body, file_name)
				if err != nil {
					return err
				}

				continue

			} else {
				lineOffset := 0
				for b_index, block := range diffFile.Blocks {
					file_name := fmt.Sprintf("%s/tmp_folder/%d_%d_%s", atom_path, file_index, b_index, filepath.Base(diffFile.TargetFile))
					addedOffset, err := saveModifiedBlockInAtom(insertBlockScriptData{
						BlockStartLine: block.OldStartLine + lineOffset - 1,
						BlockContent:   block.Body,
						Filename:       file_name,
						TargetFileName: diffFile.TargetFile,
					},
						tmp_folder_script_file,
						!cmd.Bool("block_pattern_anything"),
						cmd.Bool("add_uuid_comments"),
					)
					if err != nil {
						return err
					}
					if addedOffset {
						lineOffset++
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
	defer file.Close() //nolint:errcheck

	_, err = file.WriteString(body)
	if err != nil {
		return fmt.Errorf("could not write to output file '%s': %w", file_name, err)
	}

	return nil
}

// Saves a modified block in an atom and creates a template
func saveModifiedBlockInAtom(data insertBlockScriptData, tmp_folder_script_file *os.File, match_only_comments bool, add_uuid_comments bool) (bool, error) {
	err := os.MkdirAll(filepath.Dir(data.Filename), 0755)
	if err != nil {
		return false, fmt.Errorf("could not create output file directory for '%s': %w", data.Filename, err)
	}

	targetFileContent, err := os.ReadFile(data.TargetFileName)
	if err != nil {
		return false, fmt.Errorf("could not read target file '%s': %w", data.TargetFileName, err)
	}

	pattern, err := getLineContent(string(targetFileContent), data.BlockStartLine+1)
	if err != nil {
		pattern = ""
	}

	pattern = strings.Trim(pattern, "\n \t")
	addedComment := false
	if pattern == "" || (match_only_comments && !strings.HasPrefix(pattern, getCommentByExtension(filepath.Ext(data.TargetFileName)))) {
		pattern = "_____"
		if add_uuid_comments {
			pattern = getCommentByExtension(filepath.Ext(data.TargetFileName)) + " ! " + uuid.NewString()
			insertTextAtFileLine(pattern, data.TargetFileName, data.BlockStartLine+1, true)
			addedComment = true
		}
	}

	file, err := os.OpenFile(data.Filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return addedComment, fmt.Errorf("could not create output block file '%s': %w", data.Filename, err)
	}
	defer file.Close() //nolint:errcheck

	_, err = file.WriteString(data.BlockContent)
	if err != nil {
		return addedComment, fmt.Errorf("could not write to output file '%s': %w", data.Filename, err)
	}

	err = getInsertBlockTemplate().Execute(tmp_folder_script_file, insertBlockData{
		SearchPattern: pattern,
		BlockFile:     "tmp_folder/" + filepath.Base(data.Filename),
		TargetFile:    data.TargetFileName,
	})
	if err != nil {
		return addedComment, fmt.Errorf("could not insert block in script file: %w", err)
	}

	return addedComment, nil
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
