---
name: "kubectl-sgmap - EKS Security Groups for Pods Plugin"
description: "A kubectl plugin for displaying ENI and security group mappings in EKS clusters with Security Groups for Pods enabled. Provides comprehensive pod-to-network security auditing capabilities."
category: "DevOps & Infrastructure"
author: "naka-gawa"
authorUrl: "https://github.com/naka-gawa"
tags:
  [
    "kubectl",
    "kubernetes",
    "eks",
    "aws",
    "security-groups",
    "networking",
    "cli",
    "plugin",
    "go",
    "devops",
  ]
lastUpdated: "2025-08-25"
---

# kubectl-sgmap - EKS Security Groups for Pods Plugin

## Project Overview

`kubectl-sgmap` is a custom kubectl plugin designed specifically for Amazon EKS (Elastic Kubernetes Service) environments that have Security Groups for Pods enabled. This plugin provides visibility into the mapping between pods and their associated ENIs (Elastic Network Interfaces) and security groups, making it an essential tool for network security auditing and compliance in Kubernetes clusters.

The plugin helps DevOps engineers, platform teams, and security professionals understand and audit the network security posture of their Kubernetes workloads by clearly displaying which security groups are applied to each pod through the ENI assignments.

## Tech Stack

- **Language**: Go 1.24+
- **Framework**: Cobra CLI for command structure
- **Kubernetes Integration**: 
  - k8s.io/client-go v0.33.4
  - k8s.io/cli-runtime v0.33.4
  - k8s.io/api v0.33.4
- **AWS Integration**: AWS SDK for Go v2
  - aws-sdk-go-v2 v1.38.1
  - aws-sdk-go-v2/service/ec2 v1.245.2
- **Output Formats**: JSON, YAML, Table (default)
- **Testing**: testify v1.11.0
- **Build System**: Makefile, GoReleaser
- **Package Management**: Go Modules
- **Distribution**: kubectl krew plugin manager

## Development Environment Setup

### Prerequisites

- Go 1.25 or later
- kubectl v1.31 or newer
- AWS CLI configured with appropriate permissions
- Access to an EKS cluster with Security Groups for Pods enabled
- Docker (optional, for containerized development)

### Installation Requirements

```bash
# System requirements
go version  # Should show 1.24+
kubectl version --client  # Should show v1.31+
aws --version  # AWS CLI should be configured

# Required AWS permissions
# - ec2:DescribeNetworkInterfaces
# - ec2:DescribeSecurityGroups
# - eks:DescribeCluster (if cluster discovery is needed)
```

### Development Setup

```bash
# Clone the repository
git clone https://github.com/naka-gawa/kubectl-sgmap.git
cd kubectl-sgmap

# Install dependencies
go mod download

# Verify the build
make build

# Run tests
make test

# Install locally for development
make install
```

## Project Structure

```
kubectl-sgmap/
├── cmd/                          # Command definitions and CLI structure
│   ├── root.go                   # Root command and version handling
│   ├── sgmap.go                  # Main sgmap command implementation
│   └── pod.go                    # Pod subcommand implementation
├── internal/                     # Internal application logic
│   └── usecase/                  # Business logic layer
│       ├── pod.go                # Pod use case implementation
│       └── pod_test.go           # Pod use case tests
├── pkg/                          # Reusable packages
│   ├── aws/                      # AWS service integrations
│   │   ├── ec2.go               # EC2 service client and operations
│   │   └── ec2_test.go          # EC2 service tests
│   ├── kubernetes/               # Kubernetes client and operations
│   │   ├── client.go            # K8s client configuration
│   │   ├── pod.go               # Pod operations and queries
│   │   └── pod_test.go          # Pod operations tests
│   ├── output/                   # Output formatting utilities
│   │   └── formatter.go         # JSON, YAML, table formatters
│   └── utils/                    # Common utilities
│       ├── error.go             # Error handling utilities
│       └── config.go            # Configuration helpers
├── .github/                      # GitHub workflows and templates
│   └── workflows/               # CI/CD pipelines
│       ├── release.yml          # Release automation
│       └── test.yml             # Test automation
├── docs/                         # Documentation
│   └── RELEASING.md             # Release process documentation
├── plugins/                      # Krew plugin manifests
│   └── sgmap.yaml               # Krew plugin definition
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── main.go                       # Application entry point
├── Makefile                      # Build automation
├── .goreleaser.yaml             # Release configuration
├── .releaserc.json              # Semantic release configuration
├── aqua.yaml                    # Tool version management
└── README.md                    # Project documentation
```

## Core Architecture Principles

### Command Structure

The plugin follows kubectl's standard command pattern with a hierarchical structure:

```go
// cmd/root.go - Root command setup
package cmd

import (
    "os"
    
    "github.com/spf13/cobra"
    "k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
    version  string
    revision string
)

var streams = genericclioptions.IOStreams{
    In:     os.Stdin,
    Out:    os.Stdout,
    ErrOut: os.Stderr,
}

// Execute executes the root command
func Execute() error {
    rootCmd := NewSgmapCommand(&streams)
    rootCmd.AddCommand(newVersionCommand())
    
    return rootCmd.Execute()
}

// SetVersionInfo sets the version and revision information
func SetVersionInfo(v, r string) {
    version = v
    revision = r
}

// newVersionCommand creates the version command
func newVersionCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "version",
        Short: "Print the version number of kubectl-sgmap",
        Run: func(cmd *cobra.Command, args []string) {
            cmd.Printf("kubectl-sgmap version %s, revision %s\n", version, revision)
        },
    }
}
```

### Pod Command Implementation

```go
// cmd/pod.go - Pod subcommand with comprehensive flag support
package cmd

import (
    "context"
    "fmt"

    "github.com/spf13/cobra"
    "k8s.io/cli-runtime/pkg/genericclioptions"
    
    "github.com/naka-gawa/kubectl-sgmap/internal/usecase"
    "github.com/naka-gawa/kubectl-sgmap/pkg/aws"
    "github.com/naka-gawa/kubectl-sgmap/pkg/kubernetes"
    "github.com/naka-gawa/kubectl-sgmap/pkg/output"
)

// PodOptions contains the options for the pod command
type PodOptions struct {
    configFlags   *genericclioptions.ConfigFlags
    ioStreams     genericclioptions.IOStreams
    
    AllNamespaces bool
    OutputFormat  string
    PodName       string
    
    // Kubernetes and AWS clients will be initialized
    k8sClient *kubernetes.Client
    awsClient *aws.EC2Client
}

// NewPodCommand creates the pod subcommand
func NewPodCommand(streams genericclioptions.IOStreams) *cobra.Command {
    opts := &PodOptions{
        configFlags: genericclioptions.NewConfigFlags(true),
        ioStreams:   streams,
    }
    
    cmd := &cobra.Command{
        Use:     "pod [POD_NAME]",
        Aliases: []string{"pods", "po"},
        Short:   "Display security group information for pods",
        Long: \`Display ENI and security group mappings for pods in EKS clusters.

This command shows which security groups are applied to each pod through 
their associated Elastic Network Interfaces (ENIs). This is particularly 
useful for auditing network security policies in EKS clusters with 
Security Groups for Pods enabled.\`,
        Example: \`  # List security groups for all pods in current namespace
  kubectl sgmap pod

  # List security groups for all pods in specific namespace
  kubectl sgmap pod -n kube-system

  # List security groups for all pods across all namespaces
  kubectl sgmap pod --all-namespaces

  # Get security groups for a specific pod
  kubectl sgmap pod my-pod-name -n default

  # Output in JSON format
  kubectl sgmap pod -o json

  # Output in YAML format
  kubectl sgmap pod -o yaml\`,
        PreRunE: opts.Validate,
        RunE:    opts.Run,
    }
    
    // Add standard kubectl flags
    opts.configFlags.AddFlags(cmd.Flags())
    
    // Add command-specific flags
    cmd.Flags().BoolVarP(&opts.AllNamespaces, "all-namespaces", "A", false,
        "List pods across all namespaces")
    cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "table",
        "Output format: table, json, yaml")
    
    return cmd
}
```

## Use Case Implementation

### Business Logic Layer

```go
// internal/usecase/pod.go - Core business logic for pod security group mapping
package usecase

import (
    "context"
    "fmt"
    "strings"

    "github.com/naka-gawa/kubectl-sgmap/pkg/aws"
    "github.com/naka-gawa/kubectl-sgmap/pkg/kubernetes"
)

// PodSecurityGroupRequest represents the input for pod security group operations
type PodSecurityGroupRequest struct {
    Namespace     string
    PodName       string
    AllNamespaces bool
}

// PodSecurityGroupInfo represents the security group information for a pod
type PodSecurityGroupInfo struct {
    PodName          string   \`json:"pod_name"\`
    Namespace        string   \`json:"namespace"\`
    IPAddress        string   \`json:"ip_address"\`
    ENIId            string   \`json:"eni_id"\`
    SecurityGroupIds []string \`json:"security_group_ids"\`
    Status           string   \`json:"status"\`
}

// PodSecurityGroupResponse represents the output of pod security group operations
type PodSecurityGroupResponse struct {
    Pods []PodSecurityGroupInfo \`json:"pods"\`
}

// PodUseCase handles the business logic for pod operations
type PodUseCase struct {
    k8sClient *kubernetes.Client
    awsClient *aws.EC2Client
}

// NewPodUseCase creates a new PodUseCase instance
func NewPodUseCase(k8sClient *kubernetes.Client, awsClient *aws.EC2Client) *PodUseCase {
    return &PodUseCase{
        k8sClient: k8sClient,
        awsClient: awsClient,
    }
}

// GetPodSecurityGroups retrieves security group information for pods
func (uc *PodUseCase) GetPodSecurityGroups(ctx context.Context, req PodSecurityGroupRequest) (*PodSecurityGroupResponse, error) {
    // Get pods from Kubernetes
    pods, err := uc.k8sClient.GetPods(ctx, kubernetes.PodListOptions{
        Namespace:     req.Namespace,
        PodName:       req.PodName,
        AllNamespaces: req.AllNamespaces,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to list pods: %w", err)
    }

    var podInfos []PodSecurityGroupInfo
    
    for _, pod := range pods {
        // Skip pods without IP addresses
        if pod.Status.PodIP == "" {
            podInfos = append(podInfos, PodSecurityGroupInfo{
                PodName:   pod.Name,
                Namespace: pod.Namespace,
                Status:    string(pod.Status.Phase),
            })
            continue
        }

        // Get ENI information from AWS
        eniInfo, err := uc.awsClient.GetENIByPrivateIP(ctx, pod.Status.PodIP)
        if err != nil {
            // If ENI not found, it might be using cluster security group
            podInfos = append(podInfos, PodSecurityGroupInfo{
                PodName:   pod.Name,
                Namespace: pod.Namespace,
                IPAddress: pod.Status.PodIP,
                Status:    string(pod.Status.Phase),
            })
            continue
        }

        // Extract security group IDs
        var securityGroupIds []string
        for _, sg := range eniInfo.Groups {
            if sg.GroupId != nil {
                securityGroupIds = append(securityGroupIds, *sg.GroupId)
            }
        }

        podInfos = append(podInfos, PodSecurityGroupInfo{
            PodName:          pod.Name,
            Namespace:        pod.Namespace,
            IPAddress:        pod.Status.PodIP,
            ENIId:            *eniInfo.NetworkInterfaceId,
            SecurityGroupIds: securityGroupIds,
            Status:           string(pod.Status.Phase),
        })
    }

    return &PodSecurityGroupResponse{
        Pods: podInfos,
    }, nil
}
```

## AWS Integration

### EC2 Client Implementation

```go
// pkg/aws/ec2.go - AWS EC2 service integration
package aws

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Client wraps the AWS EC2 service client
type EC2Client struct {
    client *ec2.Client
}

// NewEC2Client creates a new EC2 client with default configuration
func NewEC2Client(ctx context.Context) (*EC2Client, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    return &EC2Client{
        client: ec2.NewFromConfig(cfg),
    }, nil
}

// GetENIByPrivateIP retrieves ENI information by private IP address
func (c *EC2Client) GetENIByPrivateIP(ctx context.Context, privateIP string) (*types.NetworkInterface, error) {
    input := &ec2.DescribeNetworkInterfacesInput{
        Filters: []types.Filter{
            {
                Name:   aws.String("private-ip-address"),
                Values: []string{privateIP},
            },
        },
    }

    result, err := c.client.DescribeNetworkInterfaces(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("failed to describe network interfaces: %w", err)
    }

    if len(result.NetworkInterfaces) == 0 {
        return nil, fmt.Errorf("no network interface found for IP %s", privateIP)
    }

    return &result.NetworkInterfaces[0], nil
}

// GetSecurityGroupsByIds retrieves security group details by IDs
func (c *EC2Client) GetSecurityGroupsByIds(ctx context.Context, groupIds []string) ([]types.SecurityGroup, error) {
    if len(groupIds) == 0 {
        return nil, nil
    }

    input := &ec2.DescribeSecurityGroupsInput{
        GroupIds: groupIds,
    }

    result, err := c.client.DescribeSecurityGroups(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("failed to describe security groups: %w", err)
    }

    return result.SecurityGroups, nil
}
```

## Kubernetes Integration

### Kubernetes Client Implementation

```go
// pkg/kubernetes/client.go - Kubernetes client configuration
package kubernetes

import (
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

// Client wraps the Kubernetes client
type Client struct {
    clientset kubernetes.Interface
}

// NewClient creates a new Kubernetes client
func NewClient(config *rest.Config) (*Client, error) {
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    return &Client{
        clientset: clientset,
    }, nil
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() kubernetes.Interface {
    return c.clientset
}
```

```go
// pkg/kubernetes/pod.go - Pod operations
package kubernetes

import (
    "context"
    "fmt"

    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodListOptions contains options for listing pods
type PodListOptions struct {
    Namespace     string
    PodName       string
    AllNamespaces bool
}

// GetPods retrieves pods based on the provided options
func (c *Client) GetPods(ctx context.Context, opts PodListOptions) ([]corev1.Pod, error) {
    if opts.PodName != "" {
        // Get specific pod
        pod, err := c.clientset.CoreV1().Pods(opts.Namespace).Get(ctx, opts.PodName, metav1.GetOptions{})
        if err != nil {
            return nil, fmt.Errorf("failed to get pod %s: %w", opts.PodName, err)
        }
        return []corev1.Pod{*pod}, nil
    }

    // List pods
    var namespace string
    if !opts.AllNamespaces {
        namespace = opts.Namespace
    }

    podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to list pods: %w", err)
    }

    return podList.Items, nil
}
```

## Output Formatting

### Multi-format Output Implementation

```go
// pkg/output/formatter.go - Output formatting utilities
package output

import (
    "encoding/json"
    "fmt"
    "io"
    "strings"
    "text/tabwriter"

    "gopkg.in/yaml.v3"
    
    "github.com/naka-gawa/kubectl-sgmap/internal/usecase"
)

// Formatter handles different output formats
type Formatter struct {
    format string
}

// NewFormatter creates a new formatter with the specified format
func NewFormatter(format string) *Formatter {
    return &Formatter{
        format: format,
    }
}

// Print outputs the data in the specified format
func (f *Formatter) Print(w io.Writer, data *usecase.PodSecurityGroupResponse) error {
    switch f.format {
    case "json":
        return f.printJSON(w, data)
    case "yaml":
        return f.printYAML(w, data)
    case "table":
        return f.printTable(w, data)
    default:
        return fmt.Errorf("unsupported output format: %s", f.format)
    }
}

// printJSON outputs data in JSON format
func (f *Formatter) printJSON(w io.Writer, data *usecase.PodSecurityGroupResponse) error {
    encoder := json.NewEncoder(w)
    encoder.SetIndent("", "  ")
    return encoder.Encode(data)
}

// printYAML outputs data in YAML format
func (f *Formatter) printYAML(w io.Writer, data *usecase.PodSecurityGroupResponse) error {
    encoder := yaml.NewEncoder(w)
    defer encoder.Close()
    return encoder.Encode(data)
}

// printTable outputs data in table format
func (f *Formatter) printTable(w io.Writer, data *usecase.PodSecurityGroupResponse) error {
    tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
    defer tw.Flush()

    // Header
    fmt.Fprintln(tw, "POD NAME\tNAMESPACE\tIP ADDRESS\tENI ID\tSECURITY GROUP IDS\tSTATUS")

    // Rows
    for _, pod := range data.Pods {
        eniId := pod.ENIId
        if eniId == "" {
            eniId = "-"
        }

        ipAddress := pod.IPAddress
        if ipAddress == "" {
            ipAddress = "-"
        }

        securityGroups := strings.Join(pod.SecurityGroupIds, ",")
        if securityGroups == "" {
            securityGroups = "-"
        }

        fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
            pod.PodName,
            pod.Namespace,
            ipAddress,
            eniId,
            securityGroups,
            pod.Status,
        )
    }

    return nil
}
```

## Testing Implementation

### Unit Testing with Table-Driven Tests

```go
// internal/usecase/pod_test.go - Unit tests for pod use case
package usecase

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    "github.com/naka-gawa/kubectl-sgmap/pkg/aws"
    "github.com/naka-gawa/kubectl-sgmap/pkg/kubernetes"
)

// MockK8sClient is a mock implementation of kubernetes.Client
type MockK8sClient struct {
    mock.Mock
}

func (m *MockK8sClient) GetPods(ctx context.Context, opts kubernetes.PodListOptions) ([]corev1.Pod, error) {
    args := m.Called(ctx, opts)
    return args.Get(0).([]corev1.Pod), args.Error(1)
}

// MockAWSClient is a mock implementation of aws.EC2Client
type MockAWSClient struct {
    mock.Mock
}

func (m *MockAWSClient) GetENIByPrivateIP(ctx context.Context, privateIP string) (*aws.NetworkInterface, error) {
    args := m.Called(ctx, privateIP)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*aws.NetworkInterface), args.Error(1)
}

func TestPodUseCase_GetPodSecurityGroups(t *testing.T) {
    tests := []struct {
        name           string
        request        PodSecurityGroupRequest
        mockPods       []corev1.Pod
        mockENI        *aws.NetworkInterface
        expectedPods   int
        expectedError  string
    }{
        {
            name: "successful retrieval with security groups",
            request: PodSecurityGroupRequest{
                Namespace: "default",
                PodName:   "",
                AllNamespaces: false,
            },
            mockPods: []corev1.Pod{
                {
                    ObjectMeta: metav1.ObjectMeta{
                        Name:      "test-pod",
                        Namespace: "default",
                    },
                    Status: corev1.PodStatus{
                        Phase: corev1.PodRunning,
                        PodIP: "192.168.1.100",
                    },
                },
            },
            mockENI: &aws.NetworkInterface{
                NetworkInterfaceId: aws.String("eni-123456789"),
                Groups: []aws.GroupIdentifier{
                    {
                        GroupId: aws.String("sg-123456789"),
                    },
                },
            },
            expectedPods: 1,
        },
        {
            name: "pod without IP address",
            request: PodSecurityGroupRequest{
                Namespace: "default",
            },
            mockPods: []corev1.Pod{
                {
                    ObjectMeta: metav1.ObjectMeta{
                        Name:      "pending-pod",
                        Namespace: "default",
                    },
                    Status: corev1.PodStatus{
                        Phase: corev1.PodPending,
                        PodIP: "",
                    },
                },
            },
            expectedPods: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockK8s := new(MockK8sClient)
            mockAWS := new(MockAWSClient)

            mockK8s.On("GetPods", mock.Anything, mock.Anything).Return(tt.mockPods, nil)
            
            if tt.mockENI != nil {
                mockAWS.On("GetENIByPrivateIP", mock.Anything, "192.168.1.100").Return(tt.mockENI, nil)
            }

            uc := NewPodUseCase(mockK8s, mockAWS)
            
            result, err := uc.GetPodSecurityGroups(context.Background(), tt.request)

            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
                assert.Len(t, result.Pods, tt.expectedPods)
            }

            mockK8s.AssertExpectations(t)
            mockAWS.AssertExpectations(t)
        })
    }
}
```

## Build and Release Configuration

### Makefile

```makefile
# Makefile - Build automation
.PHONY: build test clean install lint fmt vet

# Variables
BINARY_NAME=kubectl-sgmap
VERSION ?= $(shell git describe --tags --always --dirty)
REVISION ?= $(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Revision=$(REVISION)"

# Build commands
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Test commands
test:
	go test -v ./...

test-coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

# Development commands
install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out coverage.html

# Code quality
lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

# Dependencies
deps:
	go mod download
	go mod tidy

# Release
release:
	goreleaser release --clean
```

### GoReleaser Configuration

```yaml
# .goreleaser.yaml - Release automation configuration
version: 2

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.Version={{.Version}} -X main.Revision={{.ShortCommit}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}
      {{- if .Arm }}v{{ .Arm }}{{ end }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

snapshot:
  name_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
```

## Usage Examples and Best Practices

### Basic Usage Scenarios

```bash
# Basic usage - List all pods in current namespace
kubectl sgmap pod

# List all pods in specific namespace
kubectl sgmap pod -n kube-system

# List all pods across all namespaces
kubectl sgmap pod --all-namespaces

# Get specific pod information
kubectl sgmap pod my-app-pod -n production

# Output in different formats
kubectl sgmap pod -o json
kubectl sgmap pod -o yaml
kubectl sgmap pod -o table  # default
```

### Expected Output Examples

```bash
# Table format output
POD NAME                  NAMESPACE    IP ADDRESS       ENI ID                  SECURITY GROUP IDS           STATUS
app-deployment-12345      default      192.168.1.100    eni-0123456789abcdef0   sg-0123456789abcdef0         Running
web-service-67890         default      192.168.1.101    eni-0123456789abcdef1   sg-0123456789abcdef1         Running
database-pod-54321        data         192.168.2.50     eni-0123456789abcdef2   sg-0123456789abcdef2,sg-...  Running
```

```json
{
  "pods": [
    {
      "pod_name": "app-deployment-12345",
      "namespace": "default",
      "ip_address": "192.168.1.100",
      "eni_id": "eni-0123456789abcdef0",
      "security_group_ids": ["sg-0123456789abcdef0"],
      "status": "Running"
    }
  ]
}
```

### Troubleshooting Common Issues

```bash
# Check if pods have security groups for pods enabled
kubectl describe pod <pod-name> | grep -i "SecurityGroupForPods"

# Verify AWS credentials and permissions
aws sts get-caller-identity
aws ec2 describe-security-groups --dry-run

# Debug connection issues
kubectl sgmap pod --v=9  # Enable verbose logging

# Check if ENI exists for pod IP
aws ec2 describe-network-interfaces --filters "Name=private-ip-address,Values=<pod-ip>"
```

## Best Practices Summary

### Development Guidelines

- **Follow kubectl plugin conventions** with standard flag handling and output formats
- **Use clean architecture** with separation of concerns between CLI, business logic, and external services
- **Implement comprehensive error handling** with meaningful error messages and proper error wrapping
- **Write thorough unit tests** with mock implementations for external dependencies
- **Use table-driven tests** for comprehensive test coverage with multiple scenarios
- **Handle edge cases gracefully** such as pods without IP addresses or ENIs

### AWS Integration Best Practices

- **Use AWS SDK v2** for better performance and modern Go idioms
- **Handle AWS API rate limits** with exponential backoff and retry logic
- **Implement proper error handling** for AWS service errors and network issues
- **Use context for timeout management** in all AWS API calls
- **Minimize AWS API calls** by batching operations where possible

### Security and Compliance

- **Follow least privilege principle** for AWS IAM permissions
- **Validate input parameters** to prevent injection attacks
- **Use secure defaults** for AWS client configuration
- **Implement proper logging** without exposing sensitive information
- **Handle credentials securely** using AWS credential chain

### Performance Optimization

- **Implement concurrent processing** for multiple pod queries
- **Use connection pooling** for AWS clients
- **Cache frequently accessed data** where appropriate
- **Optimize table output** for large datasets
- **Profile memory usage** for large cluster operations

### Plugin Distribution

- **Use semantic versioning** for releases with proper changelog
- **Support multiple platforms** (Linux, macOS, Windows)
- **Provide comprehensive documentation** with examples and troubleshooting
- **Maintain backward compatibility** when possible
- **Use automated testing** in CI/CD pipelines

## Your role

- You are an excellent SRE for this project and a member of the team.
- Please support the improvement of this portfolio site in accordance with the established principles and guidelines.
- Please propose changes in pull requests with clear commit messages.

## Security and Privacy

- Important: This repository may contain sensitive information such as `.env`` files and API keys.
- Under no circumstances should the contents of these files be output (echoed) externally.

## Creating a pull request

- When creating a pull request, be sure to follow the template below.

```markdown
### Pull Request Subject

- Follow Conventional Commits, such as `feat(scope): ...` or `fix(scope): ...`.

### Pull Request Description

#### Summary
- Briefly describe what you want to achieve with this pull request.

#### Reproduction Steps (if necessary)
- For bug fixes, describe the steps to reproduce the issue.
- 1. ...
- 2. ...

#### Review Points
- Describe any specific points you would like reviewed or any design decisions you are unsure about.
- - [ ] ...
- - [ ] ...
```
