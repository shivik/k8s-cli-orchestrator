# k8sctl - Minimal Kubernetes Orchestrator CLI

A lightweight command-line tool for managing Kubernetes resources with simplified commands built using Go, Cobra, and the official Kubernetes client library.

## Features

- Deploy: Apply Kubernetes manifests and wait for rollout completion
- Status: Show pod statuses for deployments
- Logs: Tail logs from pods
- Delete: Tear down resources defined in manifests
- Dry-run: Test commands without making actual changes
- Structured Errors: Clear, actionable error messages

## Installation

### Prerequisites

- Go 1.21 or higher
- Access to a Kubernetes cluster
- `~/.kube/config` file configured

### Build from Source

```bash
git clone <repository-url>
cd k8sctl
go build -o k8sctl ./cmd/k8sctl
```

### Install Binary

```bash
# Move to a directory in your PATH
sudo mv k8sctl /usr/local/bin/
```

## Usage

### Deploy Command

Apply a Kubernetes manifest and wait for rollout completion:

```bash
# Deploy a manifest
k8sctl deploy deployment.yaml

# Dry-run mode (no changes applied)
k8sctl deploy deployment.yaml --dry-run
```

Features:
- Automatically creates or updates resources
- Waits for Deployment rollouts to complete
- Supports all Kubernetes resource types
- Uses default namespace if not specified in manifest

### Status Command

Show pod statuses for a deployment:

```bash
# Check status in default namespace
k8sctl status my-deployment

# Check status in specific namespace
k8sctl status my-deployment -n production
```

Output includes:
- Deployment replica counts (desired, updated, available, unavailable)
- Pod names, statuses, restart counts, and ages
- Formatted table view for easy reading

### Logs Command

Tail logs from a pod:

```bash
# Get last 100 lines (default)
k8sctl logs my-pod

# Get logs from specific namespace
k8sctl logs my-pod -n production

# Follow logs in real-time
k8sctl logs my-pod -f

# Get last 50 lines
k8sctl logs my-pod --tail 50
```

Features:
- Automatically selects first container if pod has multiple containers
- Supports log following (tail -f behavior)
- Configurable number of lines to display

### Delete Command

Delete resources defined in a manifest:

```bash
# Delete resources
k8sctl delete deployment.yaml

# Dry-run mode (no resources deleted)
k8sctl delete deployment.yaml --dry-run
```

Features:
- Uses foreground deletion propagation
- Waits for resources to be fully deleted
- Supports all Kubernetes resource types

## Configuration

### Kubeconfig

k8sctl reads your Kubernetes configuration from:
1. `$KUBECONFIG` environment variable (if set)
2. `~/.kube/config` (default location)

```bash
# Use custom kubeconfig
export KUBECONFIG=/path/to/custom/kubeconfig
k8sctl status my-deployment
```

### Namespaces

- Commands use the namespace specified in the manifest (for deploy/delete)
- Status and logs commands default to `default` namespace
- Override with `-n` or `--namespace` flag

## Examples

### Complete Workflow

```bash
# 1. Deploy an application
k8sctl deploy app-deployment.yaml

# 2. Check deployment status
k8sctl status my-app -n production

# 3. View application logs
k8sctl logs my-app-pod-xyz -n production -f

# 4. Clean up resources
k8sctl delete app-deployment.yaml
```

### Sample Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
```

## Error Handling

k8sctl provides structured error messages with context:

```bash
# Example error output
Error: failed to get deployment 'my-app' in namespace 'production': deployments.apps "my-app" not found

# Example validation error
Error: failed to read manifest file 'missing.yaml': open missing.yaml: no such file or directory
```

All errors include:
- Clear description of what failed
- Relevant context (resource names, namespaces)
- Underlying error details
- No silent failures

## Architecture

### Project Structure

```
k8sctl/
├── cmd/
│   └── k8sctl/
│       └── main.go          # CLI entry point
├── pkg/
│   ├── commands/
│   │   ├── deploy.go        # Deploy command implementation
│   │   ├── status.go        # Status command implementation
│   │   ├── logs.go          # Logs command implementation
│   │   └── delete.go        # Delete command implementation
│   └── k8s/
│       ├── client.go        # Kubernetes client wrapper
│       └── utils.go         # Helper functions
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
└── README.md               # This file
```

### Dependencies

- **github.com/spf13/cobra**: CLI framework
- **k8s.io/client-go**: Official Kubernetes Go client
- **k8s.io/api**: Kubernetes API types
- **k8s.io/apimachinery**: Kubernetes API machinery

## Development

### Building

```bash
# Build for current platform
go build -o k8sctl ./cmd/k8sctl

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o k8sctl-linux ./cmd/k8sctl

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o k8sctl-macos ./cmd/k8sctl
```

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Adding New Commands

1. Create a new file in `pkg/commands/`
2. Implement the command using Cobra
3. Add the command to `cmd/k8sctl/main.go`
4. Update this README

## Limitations

- Single manifest per command (no multi-document YAML support yet)
- Logs command shows first container only (for multi-container pods)
- No support for custom resource definitions (CRDs) yet
- Rollout waiting only implemented for Deployments

## Future Enhancements

- Multi-document YAML support
- Container selection for logs
- CRD support
- StatefulSet and DaemonSet rollout waiting
- Port forwarding command
- Resource scaling command
- Configuration file support

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions, please open an issue on the GitHub repository.