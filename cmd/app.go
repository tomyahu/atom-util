package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func App() *cli.Command {
	commands := &cli.Command{
		Usage: "A terminal tool that says hi to you",

		Commands: []*cli.Command{
			{
				Name:  "hello",
				Usage: "prints \"Hello World\"",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Hello World")
					return nil
				},
			},
		},
	}

	return commands
}
