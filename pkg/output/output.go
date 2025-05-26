package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"

	awsSDK "github.com/aws/aws-sdk-go-v2/aws"
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
func outputYAML(w io.Writer, data []aws.PodSecurityGroupInfo) error {
	type sg struct {
		GroupId   string `yaml:"groupId"`
		GroupName string `yaml:"groupName"`
	}

	type out struct {
		PodName        string `yaml:"podName"`
		Namespace      string `yaml:"namespace"`
		ENI            string `yaml:"eni"`
		SecurityGroups []sg   `yaml:"securityGroups"`
	}

	var converted []out
	for _, d := range data {
		var groups []sg
		for _, g := range d.SecurityGroups {
			groups = append(groups, sg{
				GroupId:   awsSDK.ToString(g.GroupId),
				GroupName: awsSDK.ToString(g.GroupName),
			})
		}
		converted = append(converted, out{
			PodName:        d.Pod.Name,
			Namespace:      d.Pod.Namespace,
			ENI:            d.ENI,
			SecurityGroups: groups,
		})
	}

	b, err := yaml.Marshal(converted)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, string(b))
	return err
}

// outputTable outputs the data in table format
func outputTable(w io.Writer, results []aws.PodSecurityGroupInfo) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAMESPACE\tPOD\tENI\tINTERFACE TYPE\tSECURITY GROUP ID\tSECURITY GROUP NAME")

	for _, r := range results {
		var sgIDs []string
		var sgNames []string
		for _, sg := range r.SecurityGroups {
			sgIDs = append(sgIDs, awsSDK.ToString(sg.GroupId))
			sgNames = append(sgNames, awsSDK.ToString(sg.GroupName))
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			r.Pod.Namespace,
			r.Pod.Name,
			r.ENI,
			r.InterfaceType,
			strings.Join(sgIDs, ", "),
			strings.Join(sgNames, ", "),
		)
	}

	return tw.Flush()
}
