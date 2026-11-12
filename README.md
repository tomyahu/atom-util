# atom-util

Command Line Tool to modify files and create reusable [copier](https://github.com/copier-org/copier) atoms for projects.

It has 3 sub-commands:

* `find`: To find the line number a regex expression occurs
* `insert`: To insert a string or the contents of a file at a specific line
* `diff_to_atom`: Reads a file containing the result of a git diff and creates the base of a copier atom to do those changes

This tool alongside copier, can be used to make repetitive changes in a proyect.
For example: creating a new endpoint in a REST API, adding a new view in a Flutter Material App, etc.