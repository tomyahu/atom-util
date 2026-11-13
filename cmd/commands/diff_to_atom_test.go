package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var gitDiffExample = `diff --git a/.vscode/command_line_snippets.code-snippets b/.vscode/command_line_snippets.code-snippets
index c430dcc..5912b9d 100644
--- a/.vscode/command_line_snippets.code-snippets
+++ b/.vscode/command_line_snippets.code-snippets
@@ -12,5 +12,27 @@
             "cmd.${1|StringArg,IntArg,BoolArg,FloatArg|}(\"${2:flag_name}\")"
         ],
         "description": ""
+    },
+    "3ac2cc29-43d0-46a0-b03b-4f7180c4f047": {
+        "prefix": "s_argument_define",
+        "body": [
+            "&cli.${1|StringArg,IntArg,BoolArg,FloatArg|}{",
+            "\tName:\t  \"${2:argument_name}\",",
+            "\tUsageText: \"${3:argument_usage}\",",
+            "}"
+        ],
+        "description": ""
+    },
+    "16f92cc6-d2a8-4e3b-a17c-495c2554d137": {
+        "prefix": "s_flag_define",
+        "body": [
+            "&cli.${1|StringFlag,IntFlag,BoolFlag,FloatFlag|}{",
+            "\tName:\t\"${2:flag_name}\",",
+            "\tUsage:   \"${3:flag_usage}\",",
+            "\tValue:   ${4:default_value},",
+            "\tAliases: []string{\"${5:aliases}\"},",
+            "}"
+        ],
+        "description": ""
     }
 }
\ No newline at end of file
diff --git a/cmd/app.go b/cmd/app.go
index b2965cc..0ab5865 100644
--- a/cmd/app.go
+++ b/cmd/app.go
@@ -11,6 +11,7 @@ func App() *cli.Command {
 			commands.FindCommand,
 			commands.InsertCommand,
 			commands.DiffToAtomCommand,
+			commands.MyCommand,
 			# ! Insert commands here
 		},
 	}
@@ -11,6 +11,7 @@ func App() *cli.Command {
 			commands.FindCommand,
 			commands.InsertCommand,
 			commands.DiffToAtomCommand,
+			commands.MyCommand,
 			# ! Insert commands here
 		},
 	}
 
diff --git a/cmd/commands/my_command.go b/cmd/commands/my_command.go
new file mode 100644
index 0000000..93c8cc3
--- /dev/null
+++ b/cmd/commands/my_command.go
@@ -0,0 +1,19 @@
+package commands
+
+import (
+	"context"
+
+	"github.com/urfave/cli/v3"
+)
+
+var MyCommand = &cli.Command{
+	Name:               "{{ command_name }}",
+	Usage:              "{{ command_description }}",
+	CustomHelpTemplate: SubCommandHelpTemplate,
+	Arguments:          []cli.Argument{},
+	Flags:              []cli.Flag{},
+	Action: func(ctx context.Context, cmd *cli.Command) error {
+		// TODO: Implement Action
+		return nil
+	},
+}
`

var gitDiffBlocksExample = `@@ -13,7 +13,7 @@ import (
 )
 
 var insertBlockTemplateText = '
-LINE=$(atom-util find "{{.SearchPattern}}" cmd/app.go)
+LINE=$(atom-util find "{{.SearchPattern}}" {{.TargetFile}})
 atom-util insert -b "{{.BlockFile}}" "{{.TargetFile}}" $LINE
 '
 
@@ -176,7 +176,8 @@ func saveModifiedBlockInAtom(file_name string, target_file string, sub_match str
 
 	err = getInsertBlockTemplate().Execute(tmp_folder_script_file, insertBlockData{
 		SearchPattern: pattern,
-		BlockFile:     file_name,
+		BlockFile:     "tmp_folder/" + filepath.Base(file_name),
+		BlockFile:     "tmp_folder/" + filepath.Base(file_name),
 		TargetFile:    target_file,
 	})
 	if err != nil {
`

func TestParse(t *testing.T) {
	values, err := parseGitDiff(gitDiffExample)
	assert.NoError(t, err)
	assert.Len(t, values, 3)
}

func TestBlockParse(t *testing.T) {
	values, err := parseGitBlocks(gitDiffBlocksExample)
	assert.NoError(t, err)
	assert.Len(t, values, 2)

	value1 := values[0]
	value2 := values[1]

	assert.Equal(t, 16, value1.NewStartLine)
	assert.Equal(t, 16, value1.NewEndLine)
	assert.Equal(t, `LINE=$(atom-util find "{{.SearchPattern}}" {{.TargetFile}})`, value1.Body)

	expectedValue2Body := `		BlockFile:     "tmp_folder/" + filepath.Base(file_name),
		BlockFile:     "tmp_folder/" + filepath.Base(file_name),`
	assert.Equal(t, 179, value2.NewStartLine)
	assert.Equal(t, 180, value2.NewEndLine)
	assert.Equal(t, expectedValue2Body, value2.Body)
}
