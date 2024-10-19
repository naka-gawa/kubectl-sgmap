package main

import (
	"os"

	cmd "github.com/naka-gawa/kubectl-sg4pod/cmd/subcommand"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
