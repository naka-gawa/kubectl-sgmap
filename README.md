# kubectl-sg4pod

`kubectl-sg4pod` is a custom kubectl plugin that displays the mapping of ENIs (Elastic Network Interfaces) and security groups assigned to pods in an EKS (Elastic Kubernetes Service) environment with Security Groups for Pods enabled. This plugin helps in auditing and managing pod-to-network associations to ensure security and compliance in Kubernetes clusters.

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

To install `kubectl-sg4pod`, follow these steps:

```bash
git clone https://github.com/your-repo/kubectl-sg4pod.git
cd kubectl-sg4pod
make install
```

## Usage

Once installed, you can use the plugin with the following command:
This command will display a list of ENIs and security groups associated with each pod running in your EKS cluster.

```bash
kubectl sg4pod get-pods -n [NameSpace]
```

### Example Output

```bash
╰─ k sg4pod get-pods -n freeeops-rails
POD NAME                                             IP ADDRESS       ENI ID                  SECURITY GROUP IDS
xxxxx-123455678-12345                                192.168.1.1      eni-123456789abcdefgh   [sg-0123456789abcdefg]
xxxxx-123455678-12346                                192.168.10.9     eni-123456789abcdefgh   [sg-0123456789abcdefg]
~snip~
```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request with any improvements, bug fixes, or new features.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.
