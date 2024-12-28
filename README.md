# kubectl-tree

A kubectl plugin to visualize Kubernetes resource relationships in a tree-like format.

## Overview

`kubectl-tree` helps you understand the relationships between Kubernetes resources in your cluster. It shows a hierarchical view of:
- Workloads (Deployments, StatefulSets, DaemonSets)
- Their child resources (ReplicaSets, Pods)
- Related resources (Services, ConfigMaps, Secrets, PVCs)

### Examples

```
kubectl tree -n istio-system
```
![alt text](<CleanShot 2024-12-21 at 17.39.23.png>)

## Installation

### Prerequisites
- Access to a Kubernetes cluster
- kubectl installed

1. Download latest release from [releases](https://github.com/danielbooth-cloud/kubectl-tree/releases)
2. Move the binary to your PATH, e.g. `/usr/local/bin/kubectl-tree`
3. Run `kubectl tree`

### Building from source
Requires go 1.21 or higher
```
go mod init kubectl-tree
```

```
go mod tidy
```

```
go build -o kubectl-tree main.go
```

