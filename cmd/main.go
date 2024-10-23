package main

import (
	"os"

	cmd "github.com/naka-gawa/kubectl-sgmap/cmd/subcommand"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
