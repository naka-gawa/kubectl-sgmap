package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"gopkg.in/yaml.v2"

	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
)

// OutputPodSecurityGroups formats and outputs pod security group information
func OutputPodSecurityGroups(w io.Writer, data []aws.PodSecurityGroupInfo, format string) error {
	switch format {
	case "json":
		return outputJSON(w, data)
	case "yaml":
		return outputYAML(w, data)
	default:
		return outputTable(w, data)
	}
}

// outputJSON outputs the data in JSON format
func outputJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputYAML outputs the data in YAML format
func outputYAML(w io.Writer, data interface{}) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal to yaml: %w", err)
	}
	_, err = w.Write(bytes)
	return err
}

// outputTable outputs the data in table format
func outputTable(w io.Writer, data []aws.PodSecurityGroupInfo) error {
	tabw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tabw, "NAMESPACE\tPOD\tENI\tSECURITY GROUP ID\tSECURITY GROUP NAME")

	for _, info := range data {
		for _, sg := range info.SecurityGroups {
			fmt.Fprintf(tabw, "%s\t%s\t%s\t%s\t%s\n",
				info.Pod.Namespace,
				info.Pod.Name,
				info.ENI,
				*sg.GroupId,
				*sg.GroupName)
		}
	}

	return tabw.Flush()
}
