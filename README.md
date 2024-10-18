# kubectl-view-podsg

`kubectl-view-podsg` is a custom kubectl plugin that displays the mapping of ENIs (Elastic Network Interfaces) and security groups assigned to pods in an EKS (Elastic Kubernetes Service) environment with Security Groups for Pods enabled. This plugin helps in auditing and managing pod-to-network associations to ensure security and compliance in Kubernetes clusters.

## Features

- Lists ENIs and security groups assigned to each pod.
- Works specifically in EKS environments with Security Groups for Pods enabled.
- Provides an easy-to-read output for network security auditing.

## Requirements

- Kubernetes version: >= 1.30
- EKS environment with Security Groups for Pods enabled
- kubectl: >= 1.31.1
- AWS CLI configured with necessary permissions

## Installation

To install `kubectl-view-podsg`, follow these steps:

```bash
git clone https://github.com/your-repo/kubectl-view-podsg.git
cd kubectl-view-podsg
make
chmod +x kubectl-view-podsg
mv kubectl-view-podsg /usr/local/bin/
```

## Usage

Once installed, you can use the plugin with the following command:
This command will display a list of ENIs and security groups associated with each pod running in your EKS cluster.

```bash
kubectl view-podsg
```

### Example Output

```bash
POD NAME          ENI ID            SECURITY GROUPS
pod-1             eni-01234abcd      sg-01234abcd
pod-2             eni-56789efgh      sg-56789efgh
```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request with any improvements, bug fixes, or new features.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.
