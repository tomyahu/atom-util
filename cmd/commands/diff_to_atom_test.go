package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var gitDiffExample = `a89ba56 (HEAD -> master) test
diff --git a/cmd/app.go b/cmd/app.go
index 67adf1c..4b017b7 100644
--- a/cmd/app.go
+++ b/cmd/app.go
@@ -7,0 +8 @@ import (
+// goodbye
@@ -16,0 +18 @@ func App() *cli.Command {
+	// hello
@@ -18,0 +21 @@ func App() *cli.Command {
+	// hello 2
diff --git a/cmd/commands/find.go b/cmd/commands/find.go
index 2975a7a..466e01a 100644
--- a/cmd/commands/find.go
+++ b/cmd/commands/find.go
@@ -12,0 +13 @@ import (
+// hello
@@ -80,0 +82 @@ var FindCommand = &cli.Command{
+			// hello 2
`

var gitDiffBlocksExample = `@@ -7,0 +8 @@ import (
+// goodbye
@@ -16,0 +18 @@ func App() *cli.Command {
+	// hello
@@ -18,0 +21,2 @@ func App() *cli.Command {
+	// hello 2
+	// hello 2
`

var gitDiffBlocksExample2 = `@@ -7,0 +8 @@ import (
+// goodbye
@@ -16,0 +18 @@ func App() *cli.Command {
+	// hello
@@ -18,0 +21,5 @@ func App() *cli.Command {
+	// hello 2
 	
+	// hello 2
 	
+	// hello 2
`

var gitDiffBlocksExample3 = `@@ -11,9 +11,11 @@ func App() *cli.Command {
 			commands.FindCommand,
 			commands.InsertCommand,
 			commands.DiffToAtomCommand,
+			commands.MyCommandCommand,
 			// ! Register commands
 		},
 	}
+	// Adding a comment here
 
 	return commands
 }`

func TestParse(t *testing.T) {
	values, err := parseGitDiff(gitDiffExample)
	assert.NoError(t, err)
	assert.Len(t, values, 2)
}

func TestBlockParse(t *testing.T) {
	values, err := parseGitBlocks(gitDiffBlocksExample)
	assert.NoError(t, err)
	assert.Len(t, values, 3)

	value1 := values[0]
	value2 := values[1]
	value3 := values[2]

	assert.Equal(t, 8, value1.NewStartLine)
	assert.Equal(t, 9, value1.NewEndLine)
	assert.Equal(t, 7, value1.OldStartLine)

	assert.Equal(t, 18, value2.NewStartLine)
	assert.Equal(t, 19, value2.NewEndLine)
	assert.Equal(t, 16, value2.OldStartLine)

	assert.Equal(t, 21, value3.NewStartLine)
	assert.Equal(t, 23, value3.NewEndLine)
	assert.Equal(t, 18, value3.OldStartLine)
}

func TestBlockParseSubBlocks(t *testing.T) {
	values, err := parseGitBlocks(gitDiffBlocksExample2)
	assert.NoError(t, err)
	assert.Len(t, values, 5)

	value1 := values[0]
	value2 := values[1]
	value3 := values[2]
	value4 := values[3]
	value5 := values[4]

	assert.Equal(t, 8, value1.NewStartLine)
	assert.Equal(t, 9, value1.NewEndLine)
	assert.Equal(t, 7, value1.OldStartLine)

	assert.Equal(t, 18, value2.NewStartLine)
	assert.Equal(t, 19, value2.NewEndLine)
	assert.Equal(t, 16, value2.OldStartLine)

	assert.Equal(t, 21, value3.NewStartLine)
	assert.Equal(t, 22, value3.NewEndLine)
	assert.Equal(t, 18, value3.OldStartLine)

	assert.Equal(t, 23, value4.NewStartLine)
	assert.Equal(t, 24, value4.NewEndLine)
	assert.Equal(t, 19, value4.OldStartLine)

	assert.Equal(t, 25, value5.NewStartLine)
	assert.Equal(t, 26, value5.NewEndLine)
	assert.Equal(t, 20, value5.OldStartLine)
}

func TestBlockParseBigBlock(t *testing.T) {
	values, err := parseGitBlocks(gitDiffBlocksExample3)
	assert.NoError(t, err)
	assert.Len(t, values, 2)

	value1 := values[0]
	value2 := values[1]

	assert.Equal(t, 14, value1.OldStartLine)
	assert.Equal(t, 17, value2.OldStartLine)
}
