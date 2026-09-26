package main

import (
	"context"
	"fmt"
	"os"

	"spx/internal/applicationcommands"
)

func main() {
	command := applicationcommands.NewRootCommand()
	command.SetOut(os.Stdout)
	command.SetErr(os.Stderr)
	if err := command.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
