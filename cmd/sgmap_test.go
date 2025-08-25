package cmd

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestNewSgmapCommand(t *testing.T) {
	streams := &genericclioptions.IOStreams{
		In:     bytes.NewBufferString(""),
		Out:    io.Discard,
		ErrOut: io.Discard,
	}
	cmd := NewSgmapCommand(streams)

	assert.NotNil(t, cmd)
	assert.Equal(t, "sgmap", cmd.Use)
	assert.True(t, cmd.HasSubCommands())

	// Check if pod command is added
	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "pod" {
			found = true
			break
		}
	}
	assert.True(t, found, "pod subcommand should be added")
}

func TestSgmapCommand_Help(t *testing.T) {
	streams := &genericclioptions.IOStreams{
		In:     bytes.NewBufferString(""),
		Out:    io.Discard,
		ErrOut: io.Discard,
	}
	cmd := NewSgmapCommand(streams)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "A kubectl plugin to display security group information for Kubernetes workloads")
	assert.Contains(t, output, "A kubectl plugin to display security group information for Kubernetes workloads running on AWS")
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "sgmap [command]")
}
