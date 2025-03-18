package utils

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

// OutputYAML は YAML形式で出力する関数です
func OutputYAML(data interface{}, out io.Writer) error {
	if err := validatePodInfo(data); err != nil {
		return err
	}

	yamlOutput, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}
	_, err = fmt.Fprintf(out, "%s", yamlOutput)
	return err
}
