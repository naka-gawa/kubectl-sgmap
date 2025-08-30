package cmd

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestNewPodCommand(t *testing.T) {
	streams := &genericclioptions.IOStreams{
		In:     bytes.NewBufferString(""),
		Out:    io.Discard,
		ErrOut: io.Discard,
	}
	cmd := NewPodCommand(streams)

	assert.NotNil(t, cmd)
	assert.Equal(t, "pod [NAME]", cmd.Use)
	assert.Equal(t, []string{"pods", "po"}, cmd.Aliases)

	// Test flags
	assert.NotNil(t, cmd.Flag("output"))
	assert.NotNil(t, cmd.Flag("all-namespaces"))
	assert.NotNil(t, cmd.Flag("namespace")) // from ConfigFlags
}

func TestPodCommand_Args(t *testing.T) {
	streams := &genericclioptions.IOStreams{
		In:     bytes.NewBufferString(""),
		Out:    io.Discard,
		ErrOut: new(bytes.Buffer),
	}

	cmd := NewPodCommand(streams)
	
	// Test argument validation only (without executing the command)
	t.Run("accepts zero arguments", func(t *testing.T) {
		err := cmd.Args(cmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("accepts one argument", func(t *testing.T) {
		err := cmd.Args(cmd, []string{"pod-name"})
		assert.NoError(t, err)
	})

	t.Run("rejects too many arguments", func(t *testing.T) {
		err := cmd.Args(cmd, []string{"pod1", "pod2"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "accepts at most 1 arg(s), received 2")
	})
}

func TestPodCommand_Help(t *testing.T) {
	streams := &genericclioptions.IOStreams{
		In:     bytes.NewBufferString(""),
		Out:    io.Discard,
		ErrOut: io.Discard,
	}
	cmd := NewPodCommand(streams)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Display security group information for pods")
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "pod [NAME] [flags]")
}
