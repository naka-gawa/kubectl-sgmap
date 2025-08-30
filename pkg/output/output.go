// Package output provides functions for formatting and outputting pod security group information.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"

	awsSDK "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
)

// OutputPodSecurityGroups formats and outputs pod security group information
func OutputPodSecurityGroups(w io.Writer, data []aws.PodSecurityGroupInfo, format string, sortField string) error {
	sort.SliceStable(data, func(i, j int) bool {
		switch sortField {
		case "ip":
			ipA := net.ParseIP(data[i].Pod.Status.PodIP)
			ipB := net.ParseIP(data[j].Pod.Status.PodIP)
			if ipA == nil || ipB == nil {
				return data[i].Pod.Status.PodIP < data[j].Pod.Status.PodIP
			}
			return bytes.Compare(ipA, ipB) < 0
		case "eni":
			return data[i].ENI < data[j].ENI
		case "attachment":
			return data[i].AttachmentLevel < data[j].AttachmentLevel
		case "sgids":
			sgsI := make([]string, len(data[i].SecurityGroups))
			for k, sg := range data[i].SecurityGroups {
				sgsI[k] = awsSDK.ToString(sg.GroupId)
			}
			sgsJ := make([]string, len(data[j].SecurityGroups))
			for k, sg := range data[j].SecurityGroups {
				sgsJ[k] = awsSDK.ToString(sg.GroupId)
			}
			return strings.Join(sgsI, ",") < strings.Join(sgsJ, ",")
		case "pod":
			fallthrough
		default:
			return data[i].Pod.Name < data[j].Pod.Name
		}
	})

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
		GroupID   string `yaml:"groupId"`
		GroupName string `yaml:"groupName"`
	}

	type out struct {
		PodName         string `yaml:"podName"`
		Namespace       string `yaml:"namespace"`
		ENI             string `yaml:"eni"`
		AttachmentLevel string `yaml:"attachmentLevel"`
		SecurityGroups  []sg   `yaml:"securityGroups"`
	}

	var converted []out
	for _, d := range data {
		var groups []sg
		for _, g := range d.SecurityGroups {
			groups = append(groups, sg{
				GroupID:   awsSDK.ToString(g.GroupId),
				GroupName: awsSDK.ToString(g.GroupName),
			})
		}
		converted = append(converted, out{
			PodName:         d.Pod.Name,
			Namespace:       d.Pod.Namespace,
			ENI:             d.ENI,
			AttachmentLevel: d.AttachmentLevel,
			SecurityGroups:  groups,
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
	fmt.Fprintln(tw, "POD NAME\tIP ADDRESS\tENI ID\tATTACHMENT\tSECURITY GROUPS")

	for _, r := range results {
		var sgs []string
		for _, sg := range r.SecurityGroups {
			sgID := awsSDK.ToString(sg.GroupId)
			sgName := awsSDK.ToString(sg.GroupName)
			if sgName != "" {
				sgs = append(sgs, fmt.Sprintf("%s (%s)", sgID, sgName))
			} else {
				sgs = append(sgs, sgID)
			}
		}
		podIP := ""
		if r.Pod.Status.PodIP != "" {
			podIP = r.Pod.Status.PodIP
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			r.Pod.Name,
			podIP,
			r.ENI,
			r.AttachmentLevel,
			strings.Join(sgs, ", "),
		)
	}

	return tw.Flush()
}
