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
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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
	case "json-minimal":
		return outputJSONMinimal(w, data)
	case "yaml":
		return outputYAML(w, data)
	default:
		return outputTable(w, data)
	}
}

// outputJSON outputs the data in JSON format
func outputJSON(w io.Writer, data []aws.PodSecurityGroupInfo) error {
	outputData := toMinimalOutput(data)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(outputData)
}

// outputJSONMinimal outputs the data in minimal JSON format
func outputJSONMinimal(w io.Writer, data []aws.PodSecurityGroupInfo) error {
	outputData := toMinimalOutput(data)
	b, err := json.Marshal(outputData)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// toMinimalOutput converts the full pod security group info into a minimal structure for output
func toMinimalOutput(data []aws.PodSecurityGroupInfo) []PodOutput {
	output := make([]PodOutput, 0, len(data))
	for _, d := range data {
		sgs := make([]SecurityGroupOutput, 0, len(d.SecurityGroups))
		for _, sg := range d.SecurityGroups {
			sgs = append(sgs, SecurityGroupOutput{
				ID:            awsSDK.ToString(sg.GroupId),
				Name:          sg.GroupName,
				InboundRules:  toRuleOutput(sg.IpPermissions, true),
				OutboundRules: toRuleOutput(sg.IpPermissionsEgress, false),
			})
		}
		output = append(output, PodOutput{
			PodName:         d.Pod.Name,
			Namespace:       d.Pod.Namespace,
			PodIP:           d.Pod.Status.PodIP,
			ENI:             d.ENI,
			AttachmentLevel: d.AttachmentLevel,
			SecurityGroups:  sgs,
		})
	}
	return output
}

func toRuleOutput(permissions []types.IpPermission, isInbound bool) []RuleOutput {
	rules := make([]RuleOutput, 0, len(permissions))
	for _, p := range permissions {
		rule := RuleOutput{
			Protocol: awsSDK.ToString(p.IpProtocol),
			FromPort: p.FromPort,
			ToPort:   p.ToPort,
		}
		if isInbound {
			for _, ipRange := range p.IpRanges {
				rule.Sources = append(rule.Sources, awsSDK.ToString(ipRange.CidrIp))
			}
			for _, ipv6Range := range p.Ipv6Ranges {
				rule.Sources = append(rule.Sources, awsSDK.ToString(ipv6Range.CidrIpv6))
			}
			for _, userGroup := range p.UserIdGroupPairs {
				rule.Sources = append(rule.Sources, awsSDK.ToString(userGroup.GroupId))
			}
		} else {
			for _, ipRange := range p.IpRanges {
				rule.Destinations = append(rule.Destinations, awsSDK.ToString(ipRange.CidrIp))
			}
			for _, ipv6Range := range p.Ipv6Ranges {
				rule.Destinations = append(rule.Destinations, awsSDK.ToString(ipv6Range.CidrIpv6))
			}
			for _, userGroup := range p.UserIdGroupPairs {
				rule.Destinations = append(rule.Destinations, awsSDK.ToString(userGroup.GroupId))
			}
		}
		rules = append(rules, rule)
	}
	return rules
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
