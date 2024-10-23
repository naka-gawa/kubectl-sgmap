package main

import (
	"testing"

	cmd "github.com/naka-gawa/kubectl-sgmap/cmd/subcommand"
)

func TestRun(t *testing.T) {
	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
