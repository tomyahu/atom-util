package diff_to_atom

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var fileRegex = regexp.MustCompile(`\nindex [\d\w]+\.\.[\d\w]+(?: \d+)?\n--- (.*)\n\+\+\+ b/(.*)\n((:?.*\n)+?)diff`)
var blockHeaderRegex = regexp.MustCompile(`@@ -(\d+),(\d+) \+(\d+)(?:,(\d+))? .*`)
var plusRegex = regexp.MustCompile(`^\+`)

// Parses a git diff
func ParseDiffFile(diffContent string) (*GitDiff, error) {
	diffContent = strings.ReplaceAll(diffContent, "\\ No newline at end of file\n", "")
	diffFiles := []GitDiffFileData{}
	for _, fileMatch := range fileRegex.FindAllStringSubmatch(diffContent+"diff", -1) {
		groups_expected := 4
		if len(fileMatch) < groups_expected {
			return nil, fmt.Errorf("unexpected number of groups captured when evaluating the regex, expected %d groups and got %d: %s", groups_expected, len(fileMatch), fileMatch)
		}

		body := fileMatch[3]
		target_file := fileMatch[2]
		is_new_file := fileMatch[1] == "/dev/null"
		blocks, err := parseFileDiffBlocks(body)
		if err != nil {
			return nil, fmt.Errorf("could not parse git diff blocks for file '%s': %w", target_file, err)
		}

		diffFiles = append(diffFiles, GitDiffFileData{
			TargetFile: target_file,
			Blocks:     blocks,
			IsNewFile:  is_new_file,
		})
	}

	return &GitDiff{ModifiedFiles: diffFiles}, nil
}

// Parses individual change blocks for a git diff file
func parseFileDiffBlocks(gitDiffFileBody string) ([]GitDiffBlock, error) {
	var result = []GitDiffBlock{}

	rawBlocks, err := getRawFileDiffBlocks(gitDiffFileBody)
	if err != nil {
		return nil, fmt.Errorf("could not parse raw diff file blocks, they are probably malformed: %w", err)
	}

	for _, rawBlock := range rawBlocks {
		blockLines := strings.Split(rawBlock, "\n")
		headerLine := blockLines[0]
		headerMatch := blockHeaderRegex.FindStringSubmatch(headerLine)

		if len(headerMatch) != 5 {
			return nil, fmt.Errorf("parsed header line has an incorrect number of matched groups: %s", headerLine)
		}

		oldStartBlock, err := strconv.Atoi(headerMatch[1])
		if err != nil {
			return nil, fmt.Errorf("could not extract old start block: %w", err)
		}

		for _, parsedBlock := range parseBlockBody(blockLines[1:], oldStartBlock) {
			result = append(result, parsedBlock)
		}
	}

	return result, nil
}

// Parses the body of a git diff block
func parseBlockBody(blockBodyLines []string, oldBlockStart int) []GitDiffBlock {
	blocks := []GitDiffBlock{}
	previousPlusLine := false
	currentBody := []string{}
	for _, lineContent := range blockBodyLines {
		if !plusRegex.MatchString(lineContent) {
			if previousPlusLine {
				previousPlusLine = false
				blocks = append(blocks, GitDiffBlock{
					OldStartLine: oldBlockStart,
					Body:         strings.Join(currentBody, "\n"),
				})
				currentBody = []string{}
			}
			oldBlockStart++

		} else {
			lineToAdd, _ := strings.CutPrefix(lineContent, "+")
			currentBody = append(currentBody, lineToAdd)
			previousPlusLine = true
		}
	}
	if previousPlusLine {
		blocks = append(blocks, GitDiffBlock{
			OldStartLine: oldBlockStart,
			Body:         strings.Join(currentBody, "\n"),
		})
	}

	return blocks
}

// Parses a raw file diff string and returns a slice of strings containing the change blocks found
func getRawFileDiffBlocks(rawFileDiff string) ([]string, error) {
	result := []string{}

	splittedDiff := strings.Split(rawFileDiff, "\n")
	if len(splittedDiff) == 0 {
		return nil, fmt.Errorf("attempted to parse an empty file diff block")
	}

	if !blockHeaderRegex.MatchString(splittedDiff[0]) {
		return nil, fmt.Errorf("first line of block is not a header block line")
	}

	currentBlockLines := []string{splittedDiff[0]}
	for i := 1; i < len(splittedDiff); i++ {
		lineContent := splittedDiff[i]
		if blockHeaderRegex.MatchString(lineContent) {
			result = append(result, strings.Join(currentBlockLines, "\n"))
			currentBlockLines = []string{lineContent}
		} else {
			currentBlockLines = append(currentBlockLines, lineContent)
		}
	}

	if len(currentBlockLines) > 0 {
		result = append(result, strings.Join(currentBlockLines, "\n"))
	}

	return result, nil
}
