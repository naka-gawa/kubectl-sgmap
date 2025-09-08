// Package output provides functions for formatting and outputting pod security group information.
package output

// PodOutput is a slimmed-down representation of a pod's security information for JSON output.
type PodOutput struct {
	PodName         string                `json:"podName"`
	Namespace       string                `json:"namespace"`
	PodIP           string                `json:"podIP"`
	ENI             string                `json:"eni"`
	AttachmentLevel string                `json:"attachmentLevel"`
	SecurityGroups  []SecurityGroupOutput `json:"securityGroups"`
}

// SecurityGroupOutput is a slimmed-down representation of a security group for JSON output.
type SecurityGroupOutput struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}
