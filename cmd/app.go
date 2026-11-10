package cmd

import (
	"example.com/find-line/cmd/commands"
	"github.com/urfave/cli/v3"
)

func App() *cli.Command {
	commands := &cli.Command{
		Commands: []*cli.Command{
			commands.FindCommand,
			commands.InsertCommand,
		},
	}

	return commands
}
