package diff_to_atom

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

var gitDiffExample2 = `32133fb feat created migration file
diff --git a/migrations/001_my_table.yaml b/migrations/001_my_table.yaml
new file mode 100644
index 0000000..f884b19
--- /dev/null
+++ b/migrations/001_my_table.yaml
@@ -0,0 +1,9 @@
+model:
+  name: my_awesome_model
+  fields:
+    - name: myfield1
+      type: int
+    - name: myfield2
+      type: String
+    - name: myfield3
+      type: String
\ No newline at end of file
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

var gitDiffBlocksExample4 = `@@ -0,0 +1,9 @@
+model:
+  name: my_awesome_model
+  fields:
+    - name: myfield1
+      type: int
+    - name: myfield2
+      type: String
+    - name: myfield3
+      type: String
\ No newline at end of file
`

func TestParse(t *testing.T) {
	diff, err := ParseDiffFile(gitDiffExample)
	assert.NoError(t, err)
	assert.Len(t, diff.ModifiedFiles, 2)
}

func TestParseCase2(t *testing.T) {
	diff, err := ParseDiffFile(gitDiffExample2)
	assert.NoError(t, err)
	assert.Len(t, diff.ModifiedFiles, 1)
}

func TestRawFileDiffBlocks(t *testing.T) {
	result, err := getRawFileDiffBlocks(gitDiffBlocksExample2)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestBlockParse(t *testing.T) {
	values, err := parseFileDiffBlocks(gitDiffBlocksExample)
	assert.NoError(t, err)
	assert.Len(t, values, 3)

	value1 := values[0]
	value2 := values[1]
	value3 := values[2]

	assert.Equal(t, 7, value1.OldStartLine)
	assert.Equal(t, 16, value2.OldStartLine)
	assert.Equal(t, 18, value3.OldStartLine)

	assert.Equal(t, "// goodbye", value1.Body)
	assert.Equal(t, "\t// hello", value2.Body)
	assert.Equal(t, "\t// hello 2\n\t// hello 2", value3.Body)
}

func TestBlockParseSubBlocks(t *testing.T) {
	values, err := parseFileDiffBlocks(gitDiffBlocksExample2)
	assert.NoError(t, err)
	assert.Len(t, values, 5)

	value1 := values[0]
	value2 := values[1]
	value3 := values[2]
	value4 := values[3]
	value5 := values[4]

	assert.Equal(t, 7, value1.OldStartLine)
	assert.Equal(t, 16, value2.OldStartLine)
	assert.Equal(t, 18, value3.OldStartLine)
	assert.Equal(t, 19, value4.OldStartLine)
	assert.Equal(t, 20, value5.OldStartLine)
}

func TestBlockParseBigBlock(t *testing.T) {
	values, err := parseFileDiffBlocks(gitDiffBlocksExample3)
	assert.NoError(t, err)
	assert.Len(t, values, 2)

	value1 := values[0]
	value2 := values[1]

	assert.Equal(t, 14, value1.OldStartLine)
	assert.Equal(t, 17, value2.OldStartLine)
}

func TestBlockParseBlockWithHyphens(t *testing.T) {
	values, err := parseFileDiffBlocks(gitDiffBlocksExample4)
	assert.NoError(t, err)
	assert.Len(t, values, 1)

	value := values[0]

	expected := `model:
  name: my_awesome_model
  fields:
    - name: myfield1
      type: int
    - name: myfield2
      type: String
    - name: myfield3
      type: String`
	assert.Equal(t, expected, value.Body)
}
