package main

import (
	"context"
	"log"
	"os"

	"example.com/find-line/cmd"
)

func main() {
	commands := cmd.App()

	if err := commands.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
