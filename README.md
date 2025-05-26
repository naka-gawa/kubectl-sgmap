# kubectl-sgmap

`kubectl-sgmap` is a custom kubectl plugin that displays the mapping of ENIs (Elastic Network Interfaces) and security groups assigned to pods in an EKS (Elastic Kubernetes Service) environment with Security Groups for Pods enabled. This plugin helps in auditing and managing pod-to-network associations to ensure security and compliance in Kubernetes clusters.

## Features

- Lists ENIs and security groups assigned to each pod.
- Works specifically in EKS environments with Security Groups for Pods enabled.
- Provides an easy-to-read output for network security auditing.

## Requirements

- Kubernetes version: >= 1.30
- EKS environment with Security Groups for Pods enabled
- kubectl: >= 1.30
- AWS CLI configured with necessary permissions

## Installation

Project `sgmap` is distributed as a kubectl plugin, and is available from the following ways:

1. kubectl plugin manager [krew](https://krew.sigs.k8s.io/docs/user-guide/setup/install/)
2. Using go install
3. Manual Installation (Build from source)

### krew Installation

1. Install [krew](https://krew.sigs.k8s.io/docs/user-guide/setup/install/) plugin manager

1. Adding a custom index

```bash
kubectl krew index add my-plugin https://github.com/naka-gawa/kubectl-sgmap.git
```

1. Install the plugin

```bash
kubectl krew install my-plugin/sgmap
```

### Using go install

If you have a latest version of Go installed you can build and install as follows:

```bash
go install github.com/naka-gawa/kubectl-sgmap@latest
```

### Manual Installation (Build from source code)

1. Clone the source (from GitHub).

```bash
git clone https://github.com/naka-gawa/kubectl-sgmap.git
cd kubectl-sgmap
make install
```

1. From the project's root directory, do the following:

```bash
cd kubectl-sgmap
make install
```

## Usage

Once installed, you can use the plugin with the following command:
This command will display a list of ENIs and security groups associated with each pod running in your EKS cluster.

```bash
kubectl sgmap pod -n [NameSpace]
```

### Example Output

```bash
kubectl sgmap pod -n test
POD NAME                  IP ADDRESS       ENI ID                  SECURITY GROUP IDS
xxxxx-123455678-12345     192.168.1.1      eni-123456789abcdefgh   [sg-0123456789abcdefg]
xxxxx-123455678-12346     192.168.10.9     eni-123456789abcdefgh   [sg-0123456789abcdefg]
~snip~
```

Support output options `--output` and `-o` to change the format.
The default is `table`. Other options are `json` and `yaml`.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with any improvements, bug fixes, or new features.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
