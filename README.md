# kubectl-eks

A kubectl plugin for **managing with Amazon EKS clusters**. This plugin provides convenient commands for listing, inspecting, and switching between EKS clusters and their associated resources.

<!-- vscode-markdown-toc -->
* [Installation](#Installation)
	* [Binary Installation](#BinaryInstallation)
* [Usage](#Usage)
	* [General Syntax](#GeneralSyntax)
	* [Available Commands](#AvailableCommands)
		* [`eks`](#eks)
		* [`use`](#use)
		* [`list`](#list)
		* [`nodegroups`](#nodegroups)
		* [`insights`](#insights)
		* [`updates`](#updates)
* [Examples](#Examples)
* [Prerequisites](#Prerequisites)

<!-- vscode-markdown-toc-config
	numbering=false
	autoSave=true
	/vscode-markdown-toc-config -->
<!-- /vscode-markdown-toc -->

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

## <a name='Usage'></a>Usage

The plugin provides several commands to interact with EKS clusters. Below is a detailed description of each command.

### <a name='GeneralSyntax'></a>General Syntax

```bash
kubectl eks [command] [flags]
```

### <a name='AvailableCommands'></a>Available Commands

#### <a name='eks'></a>`eks`

The main command for the plugin. If the current context is an EKS cluster, the plugin will show information about the cluster:

Example output:

```
$ kubectl eks
AWS PROFILE             AWS REGION   CLUSTER NAME         STATUS   VERSION   CREATED               ARN
profile-1               us-west-2    cluster-1            ACTIVE   1.29      2024-06-21 19:21:40   arn:aws:eks:us-west-2:123456789123:cluster/cluster-1
```

#### <a name='use'></a>`use`
Switch to a different EKS cluster by updating your kubeconfig.

```bash
kubectl eks use [cluster-name]
```

Example output:

```
$ kubectl eks use arn:aws:eks:us-west-2:123456789123:cluster/cluster-1
Switched to EKS cluster "cluster-1" in region "us-west-2" using profile "profile-1"
```

#### <a name='list'></a>`list`
List all EKS clusters in your AWS account with optional filters.

```bash
kubectl eks list [flags]
```

##### Flags

- `-c, --name-contains string`: Filter clusters whose names contain the specified string.
- `-p, --profile string`: Specify the AWS profile to use.
- `-q, --profile-contains string`: Filter clusters by profiles whose names contain the specified string.
- `-u, --refresh`: Refresh data from AWS.
- `-r, --region string`: Specify the AWS region to use.
- `-v, --version string`: Filter clusters by a specific Kubernetes version.

##### List all clusters:

```
$ kubectl eks list
AWS PROFILE   AWS REGION   CLUSTER NAME     STATUS   VERSION   CREATED               ARN
profile-1     us-west-2    cluster-1        ACTIVE   1.29      2024-06-21 19:21:40   arn:aws:eks:us-west-2:123456789123:cluster/cluster-1
profile-2     us-west-2    cluster-2        ACTIVE   1.31      2024-12-10 13:06:09   arn:aws:eks:us-west-2:456789123456:cluster/cluster-2
profile-3     us-west-2    cluster-3        ACTIVE   1.29      2024-12-07 19:17:50   arn:aws:eks:us-west-2:789123456789:cluster/cluster-3
```

##### Filter clusters by name containing a substring:

```
$ kubectl eks list --name-contains dev
AWS PROFILE   AWS REGION   CLUSTER NAME     STATUS   VERSION   CREATED               ARN
profile-1     us-west-2    dev-cluster-1    ACTIVE   1.29      2024-06-21 19:21:40   arn:aws:eks:us-west-2:123456789123:cluster/dev-cluster-1
profile-2     us-west-2    dev-cluster-2    ACTIVE   1.31      2024-12-10 13:06:09   arn:aws:eks:us-west-2:456789123456:cluster/dev-cluster-2
```

##### Filter clusters by region:

```
$ kubectl eks list --region us-east-1
AWS PROFILE   AWS REGION   CLUSTER NAME     STATUS   VERSION   CREATED               ARN
profile-1     us-east-1    cluster-1        ACTIVE   1.29      2024-06-21 19:21:40   arn:aws:eks:us-east-1:123456789123:cluster/cluster-1
profile-2     us-east-1    cluster-2        ACTIVE   1.31      2024-12-10 13:06:09   arn:aws:eks:us-east-1:456789123456:cluster/cluster-2
```

##### Filter clusters by Kubernetes version:

```
$ kubectl eks list --version 1.29
AWS PROFILE   AWS REGION   CLUSTER NAME     STATUS   VERSION   CREATED               ARN
profile-1     us-west-2    cluster-1        ACTIVE   1.29      2024-06-21 19:21:40   arn:aws:eks:us-west-2:123456789123:cluster/cluster-1
```

##### Filter clusters by AWS profile:

```
$ kubectl eks list --profile profile-1
AWS PROFILE   AWS REGION   CLUSTER NAME     STATUS   VERSION   CREATED               ARN
profile-1     us-west-2    cluster-1        ACTIVE   1.29      2024-06-21 19:21:40   arn:aws:eks:us-west-2:123456789123:cluster/cluster-1
```

##### Filter clusters by version and AWS profile:

```
$ kubectl eks list --version 1.29 --profile profile-1
AWS PROFILE   AWS REGION   CLUSTER NAME     STATUS   VERSION   CREATED               ARN
profile-1     us-west-2    cluster-1        ACTIVE   1.29      2024-06-21 19:21:40   arn:aws:eks:us-west-2:123456789123:cluster/cluster-1
```

#### <a name='nodegroups'></a>`nodegroups`
List the node groups of a specified EKS cluster.

```bash
kubectl eks nodegroups [cluster-name]
```

**Flags:**

- `-a, --ami string`: Describe an AMI.

##### List all node groups:

```
$ kubectl eks nodegroups my-cluster
NAME      CAPACITY TYPE   RELEASE VERSION         LAUNCH TEMPLATE        INSTANCE TYPE   DESIRED CAPACITY   MAX CAPACITY   MIN CAPACITY   VERSION   STATUS
default   ON_DEMAND       ami-123abc123abc123ab   lt-012345678abcde123   m6g.xlarge      3                  12             3              1.29      ACTIVE
```

##### Describe an AMI:

```
$ kubectl eks nodegroups -a ami-061686c363b654275
NAME                       ARCHITECTURE   STATE       DEPRECATION TIME
gd-al2023-eks-arm-1-31-9   arm64          available
gd-al2023-eks-x86-1-31-9   x86_64         available
```

#### <a name='insights'></a>`insights`
Retrieve insights about a specific EKS cluster.

```bash
kubectl eks insights [cluster-name]
```

##### Flags
- `show`: Show detailed information about a specific insight.

##### Show insights for a cluster:

```
$ kubectl eks insights my-cluster
ID                                     CATEGORY            STATUS    REASON
11111111-2222-3333-4444-555555555555   UPGRADE_READINESS   WARNING   Deprecated API usage detected within last 30 days and your cluster is on Kubernetes v1.30 or lower, or existing resources using deprecated APIs present in cluster.
```

##### Describe an insight:

```
$ kubectl eks insights my-cluster --show 11111111-2222-3333-4444-555555555555
Category: UPGRADE_READINESS
Status: WARNING
Description: Checks for usage of deprecated APIs that are scheduled for removal in Kubernetes v1.32. Upgrading your cluster before migrating to the updated APIs supported by v1.32 could cause application impact.
Recommendation: Update manifests and API clients to use newer Kubernetes APIs if applicable before upgrading to Kubernetes v1.32.
Additional Info:
  * EKS update cluster documentation:
      https://docs.aws.amazon.com/eks/latest/userguide/update-cluster.html
  * Kubernetes v1.32 deprecation guide:
      https://kubernetes.io/docs/reference/using-api/deprecation-guide/#v1-32
Deprecation Details:
  * "/apis/flowcontrol.apiserver.k8s.io/v1beta3/flowschemas" replaced with "/apis/flowcontrol.apiserver.k8s.io/v1/flowschemas"
    - Replacement from 1.29 to 1.32
    - Client Stats:
      * kubectl has requested 19 in the last 30 days - last requested: 2024-11-20 19:56:50 +0000 UTC
```

#### <a name='updates'></a>`updates`
Check for updates to the plugin or details about recent updates applied to a cluster.

```bash
kubectl eks updates [cluster-name]
```

Example output:

```
$ kubectl eks updates my-cluster
TYPE            STATUS       ERRORS
LoggingUpdate   Successful
```

## <a name='Examples'></a>Examples

1. **List all clusters:**
   ```bash
   kubectl eks list
   ```

2. **Filter clusters by name:**
   ```bash
   kubectl eks list --name-contains dev
   ```

3. **Filter clusters by AWS region:**
   ```bash
   kubectl eks list --region us-east-1
   ```

4. **Switch to a specific cluster:**
   ```bash
   kubectl eks use my-cluster
   ```

5. **Get insights about a cluster:**
   ```bash
   kubectl eks insights my-cluster
   ```

6. **View detailed information about a specific insight:**
   ```bash
   kubectl eks insights my-cluster --show 11111111-2222-3333-4444-555555555555
   ```

7. **List node groups for a cluster:**
   ```bash
   kubectl eks nodegroups my-cluster
   ```

8. **Check for updates:**
   ```bash
   kubectl eks updates my-cluster
   ```

## <a name='Prerequisites'></a>Prerequisites

- **AWS CLI:** Ensure that the AWS CLI is installed and configured with the appropriate credentials and region.
- **kubectl:** Ensure you have `kubectl` installed and configured.
