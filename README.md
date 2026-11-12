# atom-util

Command Line Tool to modify files and create reusable [copier](https://github.com/copier-org/copier) atoms for projects.

It has 3 sub-commands:

* `find`: To find the line number a regex expression occurs
* `insert`: To insert a string or the contents of a file at a specific line
* `diff_to_atom`: Reads a file containing the result of a git diff and creates the base of a copier atom to do those changes

This tool alongside copier, can be used to make repetitive changes in a proyect.
For example: creating a new endpoint in a REST API, adding a new view in a Flutter Material App, etc.

## Requirements

* Go 1.25
* [copier](https://github.com/copier-org/copier) if you want to use the atoms you create with `diff_to_atom` (you can just use this to check for lines and edit files if you want to though, no preassure)

## Installation

1. Clone this repository
1. Run `make install` and `make build`
1. Move the binary to a path your terminal can find it
	* In Linux and MacOS this can be
		* `~/.local/bin` if you want it to be available only for the current user
		* `/usr/local/bin` if you want it to be available for all users
	* In Windows add it to a folder that is added to the `PATH` environment variable. Check [this video](https://www.youtube.com/watch?v=F1ZRMf7gxsM) if you don't know how to do that.

