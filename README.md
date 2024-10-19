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
make
chmod +x kubectl-sg4pod
mv kubectl-sg4pod /usr/local/bin/
```

## Usage

Once installed, you can use the plugin with the following command:
This command will display a list of ENIs and security groups associated with each pod running in your EKS cluster.

```bash
kubectl sg4pod
```

### Example Output

```bash
POD NAME          ENI ID            SECURITY GROUPS
pod-1             eni-01234abcd      [sg-01234abcd]
pod-2             eni-56789efgh      [sg-56789efgh]
```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request with any improvements, bug fixes, or new features.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.
