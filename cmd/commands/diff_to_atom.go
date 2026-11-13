package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

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

type gitDiffBlock struct {
	NewStartLine int
	NewEndLine   int
	Body         string
}

type gitDiffFileData struct {
	TargetFile string
	Blocks     []gitDiffBlock
	IsNewFile  bool
}

type insertBlockScriptData struct {
	BlockStartLine int
	BlockEndLine   int
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
		parsedGitDiff, err := parseGitDiff(string(data))
		if err != nil {
			return fmt.Errorf("failed to parse git diff: %w", err)
		}

		for file_index, diffFile := range parsedGitDiff {

			if diffFile.IsNewFile {
				file_name := atom_path + "/" + diffFile.TargetFile
				err = saveNewFileInAtom(diffFile.Blocks[0].Body, file_name)
				if err != nil {
					return err
				}

				continue
			} else {

				for b_index, block := range diffFile.Blocks {
					file_name := fmt.Sprintf("%s/tmp_folder/%d_%d_%s", atom_path, file_index, b_index, filepath.Base(diffFile.TargetFile))
					err = saveModifiedBlockInAtom(insertBlockScriptData{
						BlockStartLine: block.NewStartLine,
						BlockEndLine:   block.NewEndLine,
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
				}
			}
		}
		return nil
	},
}

func parseGitDiff(gitDiff string) ([]gitDiffFileData, error) {
	regex := regexp.MustCompile(`\nindex [\d\w]+\.\.[\d\w]+(?: \d+)?\n--- (.*)\n\+\+\+ b/(.*)\n((:?.*\n)+?)diff`)
	var result = []gitDiffFileData{}
	gitDiff = strings.ReplaceAll(gitDiff, "\\ No newline at end of file\n", "")
	matches := regex.FindAllStringSubmatch(gitDiff+"diff", -1)
	for _, match := range matches {
		groups_expected := 4
		if len(match) < groups_expected {
			return []gitDiffFileData{}, fmt.Errorf("unexpected number of groups captured when evaluating the regex, expected %d groups and got %d: %s", groups_expected, len(match), match)
		}

		body := match[3]
		target_file := match[2]
		is_new_file := match[1] == "/dev/null"
		blocks, err := parseGitBlocks(body)
		if err != nil {
			return []gitDiffFileData{}, fmt.Errorf("could not parse git diff blocks for file '%s': %w", target_file, err)
		}

		result = append(result, gitDiffFileData{
			TargetFile: target_file,
			Blocks:     blocks,
			IsNewFile:  is_new_file,
		})

	}

	return result, nil
}

// Parses individual change blocks for a git diff file
func parseGitBlocks(gitDiffFileBody string) ([]gitDiffBlock, error) {
	headerRegex := regexp.MustCompile(`@@ -(\d+),(\d+) \+(\d+),(\d+) .*`)
	plus_regex := regexp.MustCompile(`^\+`)
	minus_regex := regexp.MustCompile(`^\-`)

	splittedBody := strings.Split(gitDiffFileBody, "\n")

	var result = []gitDiffBlock{}
	i := 0
	for i < len(splittedBody)-2 {
		lineContent := splittedBody[i]
		if !headerRegex.MatchString(lineContent) {
			return []gitDiffBlock{}, fmt.Errorf("parsed header line is not valid: %s", lineContent)
		}

		match := headerRegex.FindStringSubmatch(lineContent)

		if len(match) != 5 {
			return []gitDiffBlock{}, fmt.Errorf("parsed header line has an incorrect number of matched groups: %s", lineContent)
		}

		newStartBlock, err := strconv.Atoi(match[3])
		if err != nil {
			return []gitDiffBlock{}, fmt.Errorf("could not extract new start block: %w", err)
		}

		blockExtends, err := strconv.Atoi(match[4])
		if err != nil {
			return []gitDiffBlock{}, fmt.Errorf("could not extract new block line amount: %w", err)
		}

		i++
		j := 0
		startBlockLine := -1
		endBlockLine := -1
		blockBody := []string{}
		for j < blockExtends {
			lineContent := splittedBody[i]
			if plus_regex.MatchString(lineContent) && startBlockLine == -1 {
				startBlockLine = j + newStartBlock
			}
			if !minus_regex.MatchString(lineContent) && !plus_regex.MatchString(lineContent) && startBlockLine != -1 && endBlockLine == -1 {
				endBlockLine = j + newStartBlock - 1
			}
			if plus_regex.MatchString(lineContent) {
				lineToAdd, _ := strings.CutPrefix(lineContent, "+")
				blockBody = append(blockBody, lineToAdd)
			}
			if !minus_regex.MatchString(lineContent) {
				j++
			}
			i++
		}

		result = append(result, gitDiffBlock{
			NewStartLine: startBlockLine,
			NewEndLine:   endBlockLine,
			Body:         strings.Join(blockBody, "\n"),
		})

	}

	return result, nil
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
func saveModifiedBlockInAtom(data insertBlockScriptData, tmp_folder_script_file *os.File, match_only_comments bool, add_uuid_comments bool) error {
	err := os.MkdirAll(filepath.Dir(data.Filename), 0755)
	if err != nil {
		return fmt.Errorf("could not create output file directory for '%s': %w", data.Filename, err)
	}

	targetFileContent, err := os.ReadFile(data.TargetFileName)
	if err != nil {
		return fmt.Errorf("could not read target file '%s': %w", data.TargetFileName, err)
	}

	pattern, err := getLineContent(string(targetFileContent), data.BlockEndLine)
	if err != nil {
		pattern = ""
	}

	pattern = strings.Trim(pattern, "\n \t")
	if pattern == "" || (match_only_comments && !strings.HasPrefix(pattern, "//") && !strings.HasPrefix(pattern, "#")) {
		pattern = "_____"
		if add_uuid_comments {
			pattern = "// ! " + uuid.NewString()
			insertTextAtFileLine(pattern, data.TargetFileName, data.BlockEndLine, true)
		}
	}

	file, err := os.OpenFile(data.Filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("could not create output block file '%s': %w", data.Filename, err)
	}
	defer file.Close() //nolint:errcheck

	_, err = file.WriteString(data.BlockContent)
	if err != nil {
		return fmt.Errorf("could not write to output file '%s': %w", data.Filename, err)
	}

	err = getInsertBlockTemplate().Execute(tmp_folder_script_file, insertBlockData{
		SearchPattern: pattern,
		BlockFile:     "tmp_folder/" + filepath.Base(data.Filename),
		TargetFile:    data.TargetFileName,
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
