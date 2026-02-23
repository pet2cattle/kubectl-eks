## kubectl-eks events

Show aggregated events across namespaces

### Synopsis

Show aggregated events across namespaces, sorted by time.

Displays cluster events, with options to filter by type (Warning/Normal)
and namespace. Events are sorted by timestamp (most recent first).
Shows all event types by default.

```
kubectl-eks events [flags]
```

### Examples

```
  # Show all events across all namespaces
  kubectl eks events

  # Show only warning events
  kubectl eks events --warnings-only

  # Show events for specific namespace
  kubectl eks events -n kube-system

  # Show all events (explicit)
  kubectl eks events --all
```

### Options

```
      --all                Show all events (default behavior)
  -A, --all-namespaces     Show events across all namespaces (default)
  -h, --help               help for events
  -n, --namespace string   Namespace to show events for
      --warnings-only      Show only warning events
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                  UID to impersonate for the operation.
      --as-user-extra stringArray      User extras to impersonate for the operation, this flag can be repeated to specify multiple values for the same key.
      --cache-dir string               Default cache directory (default "/Users/jprats/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --disable-compression            If true, opt-out of response compression for all requests to the server
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
      --no-headers                     When using the default or custom-column output format, don't print headers (default print headers)
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

### SEE ALSO

* [kubectl-eks](kubectl-eks.md)	 - A kubectl plugin for managing Amazon EKS clusters

