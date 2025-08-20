# kubectl-sgmap

![Release](https://img.shields.io/github/v/release/naka-gawa/kubectl-sgmap?color=blue)
![Downloads](https://img.shields.io/github/downloads/naka-gawa/kubectl-sgmap/total?color=green)
![License](https://img.shields.io/github/license/bmcustodio/kubectl-topology)
![Stars](https://img.shields.io/github/stars/naka-gawa/kubectl-sgmap?style=social)

`kubectl-sgmap` is a custom kubectl plugin that displays the mapping of ENIs (Elastic Network Interfaces) and security groups assigned to pods in an EKS (Elastic Kubernetes Service) environment with Security Groups for Pods enabled. This plugin helps in auditing and managing pod-to-network associations to ensure security and compliance in Kubernetes clusters.

## Features

- Lists ENIs and security groups assigned to each pod.
- Works specifically in EKS environments with Security Groups for Pods enabled.
- Provides an easy-to-read output for network security auditing.

## Requirements

- **Kubernetes Version**: This plugin is built and tested against Kubernetes `v1.33`. It is expected to be compatible with Kubernetes versions `v1.31` and newer.
- **kubectl Version**: The plugin is built with client libraries from `kubectl v1.33`. It should be compatible with `kubectl` versions `v1.31` and newer.
- **EKS Environment**: Requires an EKS cluster with Security Groups for Pods enabled.
- **AWS CLI**: A configured AWS CLI with permissions to describe EC2 network interfaces and security groups.

## Installation

Project `sgmap` is distributed as a kubectl plugin, and is available from the following ways:

1. **Kubectl Plugin Manager (krew)**: Recommended for most users.
2. **Go Install**: For users who have a Go environment set up.
3. **Manual Installation**: For developers and contributors.

### Krew Installation

1. **Install krew**: Follow the [official krew installation guide](https://krew.sigs.k8s.io/docs/user-guide/setup/install/) to set up the plugin manager.

2. **Add the custom plugin index**: This command adds the `kubectl-sgmap` repository as a new plugin index for krew.
```bash
kubectl krew index add sgmap-repo https://github.com/naka-gawa/kubectl-sgmap.git
```

3. **Install the plugin**: Now you can install the `sgmap` plugin from the newly added index.
```bash
kubectl krew install sgmap-repo/sgmap
```

### Go Install

If you have a Go environment configured, you can install the plugin with the following command:
```bash
go install github.com/naka-gawa/kubectl-sgmap@latest
```

### Manual Installation (from source)

To build and install the plugin from the source code, follow these steps:

1. **Clone the repository**:
```bash
git clone https://github.com/naka-gawa/kubectl-sgmap.git
```

2. **Build and install**:
```bash
cd kubectl-sgmap
make install
```
This will build the `kubectl-sgmap` binary and move it to a directory in your `$PATH`.

## Usage

The `sgmap` plugin follows the standard `kubectl` command structure.

```bash
kubectl sgmap <subcommand> [flags]
```

### Subcommands

- `pod` (aliases: `pods`, `po`): Display security group information for pods.
- `version`: Print the plugin version.

### Examples

**List security groups for all pods in the current namespace:**
```bash
kubectl sgmap pod
```

**List security groups for all pods in a specific namespace:**
```bash
kubectl sgmap pod -n <namespace>
```
*Example Output:*
```
POD NAME                  IP ADDRESS       ENI ID                  SECURITY GROUP IDS
xxxxx-123455678-12345     192.168.1.1      eni-123456789abcdefgh   [sg-0123456789abcdefg]
xxxxx-123455678-12346     192.168.10.9     eni-123456789abcdefgh   [sg-0123456789abcdefg]
~snip~
```

**List security groups for a specific pod:**
```bash
kubectl sgmap pod <pod-name> -n <namespace>
```

**List security groups for all pods in all namespaces:**
```bash
kubectl sgmap pod -A
```

**Output in JSON or YAML format:**
```bash
kubectl sgmap pod -n <namespace> -o json
```

The plugin supports all standard `kubectl` flags like `--namespace`, `--context`, and `--kubeconfig`.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with any improvements, bug fixes, or new features.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
