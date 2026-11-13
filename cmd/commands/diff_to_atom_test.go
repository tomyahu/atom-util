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

	assert.Equal(t, 18, value2.NewStartLine)
	assert.Equal(t, 19, value2.NewEndLine)

	assert.Equal(t, 21, value3.NewStartLine)
	assert.Equal(t, 23, value3.NewEndLine)
}
