# kubectl-eks

A kubectl plugin for **managing Amazon EKS clusters**. This plugin provides convenient commands for listing, inspecting, and switching between EKS clusters.

<!-- vscode-markdown-toc -->
* [Prerequisites](#Prerequisites)
* [Installation](#Installation)
	* [Binary Installation](#BinaryInstallation)
* [Documentation](#Documentation)
* [Quick Start](#QuickStart)

<!-- vscode-markdown-toc-config
	numbering=false
	autoSave=true
	/vscode-markdown-toc-config -->
<!-- /vscode-markdown-toc -->

## <a name='Prerequisites'></a>Prerequisites

- **AWS CLI:** Ensure that `aws` is installed and configured with the appropriate credentials and region.
- **kubectl:** Ensure you have `kubectl` installed and configured.

## <a name='Installation'></a>Installation

We provide pre-built binaries for Linux and macOS (arm64 and x86-64). Follow the instructions below to install the plugin on your system.

### <a name='BinaryInstallation'></a>Binary Installation

1. Download the binary for your operating system from the [releases page](https://github.com/pet2cattle/kubectl-eks/releases).

2. Install the binary:

    ```bash
    chmod +x kubectl-eks
    mv kubectl-eks /usr/local/bin/
    ```

3. Verify installation:

    ```bash
    kubectl eks --help
    ```

## <a name='Documentation'></a>Documentation

Full command reference documentation is available in the [docs/](docs/) directory:

- [kubectl eks](docs/kubectl-eks.md) - Main command and cluster information
- [kubectl eks list](docs/kubectl-eks_list.md) - List all EKS clusters
- [kubectl eks use](docs/kubectl-eks_use.md) - Switch to a different cluster
- [kubectl eks mget](docs/kubectl-eks_mget.md) - Get resources from multiple clusters
- [kubectl eks stats](docs/kubectl-eks_stats.md) - Get cluster statistics
- [kubectl eks nodegroups](docs/kubectl-eks_nodegroups.md) - List cluster node groups
- [kubectl eks insights](docs/kubectl-eks_insights.md) - Get cluster insights
- [kubectl eks updates](docs/kubectl-eks_updates.md) - Check for updates
- [kubectl eks stacks](docs/kubectl-eks_stacks.md) - Get CloudFormation stacks
- [kubectl eks aws-profile](docs/kubectl-eks_aws-profile.md) - Get AWS profile
- [kubectl eks completion](docs/kubectl-eks_completion.md) - Generate shell autocompletion

Browse all commands and their options in the [docs](docs/) folder.

## <a name='QuickStart'></a>Quick Start

```bash
# List all EKS clusters
kubectl eks list

# Filter clusters by name
kubectl eks list --name-contains prod

# Switch to a specific cluster
kubectl eks use my-cluster

# Get resources from multiple clusters
kubectl eks mget pods -q prod

# View cluster statistics
kubectl eks stats

# Get insights about a cluster
kubectl eks insights my-cluster
```