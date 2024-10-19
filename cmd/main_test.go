package main

import (
	"testing"

	cmd "github.com/naka-gawa/kubectl-sg4pod/cmd/subcommand"
)

func TestRun(t *testing.T) {
	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
