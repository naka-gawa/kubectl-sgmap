package lib

import (
	"encoding/json"
	"fmt"
	"io"
)

func OutputJSON(data interface{}, out io.Writer) error {
	if err := validatePodInfo(data); err != nil {
		return err
	}

	jsonOutput, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %v", err)
	}
	_, err = fmt.Fprintf(out, "%s", jsonOutput)
	return err
}
