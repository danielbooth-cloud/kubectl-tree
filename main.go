package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"

    "./k8s"
    "./tree"
    "./util"
    "k8s.io/client-go/util/homedir"
)

const version = "1.0.0"

func main() {
    var kubeconfig *string
    var showVersion bool
    var namespace string
    var debug bool

    if home := homedir.HomeDir(); home != "" {
        kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
    } else {
        kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
    }

    flag.BoolVar(&showVersion, "version", false, "show version information")
    flag.StringVar(&namespace, "n", "", "namespace to show tree for (defaults to current namespace)")
    flag.BoolVar(&debug, "debug", false, "enable debug output")
    flag.Parse()

    if showVersion {
        fmt.Printf("kubectl-tree version %s\n", version)
        return
    }

    // Get current namespace if not specified
    if namespace == "" {
        if ns, err := util.GetCurrentNamespace(*kubeconfig); err == nil {
            namespace = ns
        } else {
            namespace = "default"
        }
    }

    // Create kubernetes client
    client, err := k8s.NewClient(*kubeconfig)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    // Check if namespace exists
    if err := client.NamespaceExists(namespace); err != nil {
        fmt.Printf("Error: namespace '%s' not found\n", namespace)
        os.Exit(1)
    }

    // Get the tree
    root, err := tree.NewBuilder(client, debug).BuildTree(namespace)
    if err != nil {
        fmt.Printf("Error building resource tree: %v\n", os.Stderr)
        os.Exit(1)
    }

    // Create printer with color support
    printer := tree.NewPrinter(true)
    
    // Print the tree starting with empty prefix and root is the last node
    printer.PrintTree(root, "", true)
}