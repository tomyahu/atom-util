package cmd

import (
	"example.com/atom-util/cmd/commands"
	"github.com/urfave/cli/v3"
)

func App() *cli.Command {
	commands := &cli.Command{
		Commands: []*cli.Command{
			commands.FindCommand,
			commands.InsertCommand,
			commands.DiffToAtomCommand,
			// ! Register commands
		},
	}

	return commands
}
