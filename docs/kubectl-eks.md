## kubectl-eks

A kubectl plugin for managing Amazon EKS clusters

### Synopsis

kubectl-eks provides convenient commands for listing, inspecting, 
and switching between EKS clusters and their associated resources.

Prerequisites:
- AWS CLI installed and configured
- kubectl installed and configured

```
kubectl-eks [flags]
```

### Examples

```
  # Show current cluster info
  kubectl eks
  
  # List all clusters
  kubectl eks list
  
  # Switch to a cluster
  kubectl eks use my-cluster
```

### Options

```
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                  UID to impersonate for the operation.
      --cache-dir string               Default cache directory (default "/Users/jprats/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --disable-compression            If true, opt-out of response compression for all requests to the server
  -h, --help                           help for kubectl-eks
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --no-headers                     When using the default or custom-column output format, don't print headers (default print headers)
  -u, --refresh                        Do not use cached data, refresh from AWS
  -r, --region string                  Switch to the same cluster in a different region
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

### SEE ALSO

* [kubectl-eks aws-profile](kubectl-eks_aws-profile.md)	 - Get AWS profile
* [kubectl-eks completion](kubectl-eks_completion.md)	 - Generate the autocompletion script for the specified shell
* [kubectl-eks insights](kubectl-eks_insights.md)	 - Get insights about an EKS cluster
* [kubectl-eks list](kubectl-eks_list.md)	 - List all EKS clusters in your AWS account
* [kubectl-eks mget](kubectl-eks_mget.md)	 - Get resources from multiple clusters
* [kubectl-eks nodegroups](kubectl-eks_nodegroups.md)	 - List EKS nodegroups
* [kubectl-eks stacks](kubectl-eks_stacks.md)	 - Get CF stacks
* [kubectl-eks stats](kubectl-eks_stats.md)	 - Get EKS cluster stats
* [kubectl-eks updates](kubectl-eks_updates.md)	 - Check for updates
* [kubectl-eks use](kubectl-eks_use.md)	 - switch to a different EKS cluster

