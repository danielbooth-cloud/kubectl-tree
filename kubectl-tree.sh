#!/bin/bash
set -euo pipefail

# Define plugin name and version
PLUGIN_NAME="kubectl-tree"
VERSION="1.0.0"

# Help message
show_help() {
    echo "kubectl-tree - Display Kubernetes resources in a tree structure"
    echo ""
    echo "Usage:"
    echo "  kubectl tree [namespace]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -v, --version  Show version information"
    echo ""
    echo "If namespace is not specified, it uses the current namespace"
}

# Version information
show_version() {
    echo "$PLUGIN_NAME version $VERSION"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            show_version
            exit 0
            ;;
        *)
            NAMESPACE="$1"
            shift
            ;;
    esac
done

# If namespace is not provided, get current namespace
if [ -z "${NAMESPACE:-}" ]; then
    NAMESPACE=$(kubectl config view --minify --output 'jsonpath={..namespace}')
    # If still empty, use default
    NAMESPACE=${NAMESPACE:-default}
fi

# Function to print tree branches
print_tree() {
    local prefix="$1"
    local name="$2"
    echo "${prefix}${name}"
}

# Function to get resource dependencies
get_dependencies() {
    local resource_type="$1"
    local resource_name="$2"
    local prefix="$3"

    # Get owner references
    local owners=$(kubectl get "$resource_type" "$resource_name" -n "$NAMESPACE" -o jsonpath='{.metadata.ownerReferences[*].name}' 2>/dev/null)
    
    for owner in $owners; do
        # Get owner kind
        local owner_kind=$(kubectl get "$resource_type" "$resource_name" -n "$NAMESPACE" -o jsonpath="{.metadata.ownerReferences[?(@.name==\"$owner\")].kind}" | tr '[:upper:]' '[:lower:]')
        print_tree "$prefix" "└── $owner_kind/$owner"
        get_dependencies "$owner_kind" "$owner" "$prefix    "
    done
}

# Main function to build resource tree
build_tree() {
    # List of resource types to check
    local resource_types=(
        "pods"
        "deployments"
        "replicasets"
        "statefulsets"
        "daemonsets"
        "services"
        "configmaps"
        "secrets"
        "persistentvolumeclaims"
    )

    echo "Resource tree for namespace: $NAMESPACE"
    echo "├── Namespace/$NAMESPACE"

    for resource_type in "${resource_types[@]}"; do
        # Get resources of current type
        local resources=$(kubectl get "$resource_type" -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)
        
        if [ ! -z "$resources" ]; then
            for resource in $resources; do
                # Check if resource has no owner references (is a root resource)
                if [ -z "$(kubectl get "$resource_type" "$resource" -n "$NAMESPACE" -o jsonpath='{.metadata.ownerReferences}' 2>/dev/null)" ]; then
                    print_tree "│   " "└── $resource_type/$resource"
                    get_dependencies "$resource_type" "$resource" "│       "
                fi
            done
        fi
    done
}

# Execute main function
build_tree