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

func TestPodCommand_RunE(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantErr       bool
		expectedErr   string
	}{
		{
			name:          "no arguments",
			args:          []string{},
			wantErr:       true,
			expectedErr:   "connect: connection refused",
		},
		{
			name:          "with pod name",
			args:          []string{"my-pod"},
			wantErr:       true,
			expectedErr:   "connect: connection refused",
		},
		{
			name:          "with all-namespaces flag",
			args:          []string{"--all-namespaces"},
			wantErr:       true,
			expectedErr:   "connect: connection refused",
		},
		{
			name:          "with output flag",
			args:          []string{"-o", "json"},
			wantErr:       true,
			expectedErr:   "connect: connection refused",
		},
		{
			name:        "too many arguments",
			args:        []string{"pod1", "pod2"},
			wantErr:     true,
			expectedErr: "accepts at most 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streams := &genericclioptions.IOStreams{
				In:     bytes.NewBufferString(""),
				Out:    io.Discard,
				ErrOut: new(bytes.Buffer), // Capture stderr
			}

			cmd := NewPodCommand(streams)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)

				// The error from cobra goes to stderr and is not returned by Execute()
				// so we need to check both the returned error and stderr.
				stderr := streams.ErrOut.(*bytes.Buffer).String()

				// The error returned by Execute() comes from the usecase layer in valid cases
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectedErr)
				} else {
					// The error from cobra parsing goes to stderr
					assert.Contains(t, stderr, tt.expectedErr)
				}

			} else {
				assert.NoError(t, err)
			}
		})
	}
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
