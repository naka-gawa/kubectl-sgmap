package main

import (
	"os"

	cmd "github.com/naka-gawa/kubectl-view-podsg/cmd/subcommand"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
