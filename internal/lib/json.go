package lib

import (
	"encoding/json"
	"fmt"
	"io"
)

func OutputJSON(pods []PodInfo, out io.Writer) error {
	jsonOutput, err := json.MarshalIndent(pods, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %v", err)
	}
	_, err = fmt.Fprintf(out, "%s\n", jsonOutput)
	return err
}
