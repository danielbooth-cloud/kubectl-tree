package util

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const defaultNamespace = "default"

// GetCurrentNamespace returns the current namespace from the kubeconfig context
func GetCurrentNamespace(kubeconfig string) (string, error) {
	// If kubeconfig is empty, try to use the default location
	if kubeconfig == "" {
		if home := os.Getenv("HOME"); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Load the kubeconfig file
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	return getCurrentNamespaceFromConfig(config)
}

// getCurrentNamespaceFromConfig extracts the current namespace from the config
func getCurrentNamespaceFromConfig(config *api.Config) (string, error) {
	if config == nil {
		return "", fmt.Errorf("kubeconfig is nil")
	}

	// Get the current context
	context, exists := config.Contexts[config.CurrentContext]
	if !exists {
		return "", fmt.Errorf("current context %q not found in kubeconfig", config.CurrentContext)
	}

	// If namespace is set in the current context, use it
	if context.Namespace != "" {
		return context.Namespace, nil
	}

	// Default to "default" namespace if none is specified
	return defaultNamespace, nil
}

// ValidateNamespace checks if the namespace is valid
func ValidateNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}

	// Add any additional validation rules here
	// For example, check length, allowed characters, etc.
	if len(namespace) > 253 {
		return fmt.Errorf("namespace name cannot be longer than 253 characters")
	}

	return nil
}

// GetNamespaceOrDefault returns the provided namespace or gets the current namespace
func GetNamespaceOrDefault(namespace, kubeconfig string) (string, error) {
	// If namespace is provided, validate and use it
	if namespace != "" {
		if err := ValidateNamespace(namespace); err != nil {
			return "", err
		}
		return namespace, nil
	}

	// Otherwise, get the current namespace from kubeconfig
	ns, err := GetCurrentNamespace(kubeconfig)
	if err != nil {
		return defaultNamespace, nil // Fallback to default namespace
	}

	return ns, nil
}

// IsSystemNamespace returns true if the namespace is a system namespace
func IsSystemNamespace(namespace string) bool {
	systemNamespaces := map[string]bool{
		"kube-system":          true,
		"kube-public":          true,
		"kube-node-lease":      true,
		"default":              true,
		"kubernetes-dashboard": true,
	}

	return systemNamespaces[namespace]
}

// GetNamespaceFromPath extracts the namespace from a resource path
func GetNamespaceFromPath(path string) string {
	// Example path: /api/v1/namespaces/default/pods
	parts := filepath.SplitList(path)
	for i, part := range parts {
		if part == "namespaces" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}