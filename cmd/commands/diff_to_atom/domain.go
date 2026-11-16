package diff_to_atom

type GitDiffBlock struct {
	OldStartLine int
	Body         string
}

type GitDiffFileData struct {
	TargetFile string
	Blocks     []GitDiffBlock
	IsNewFile  bool
}

type GitDiff struct {
	ModifiedFiles []GitDiffFileData
}
