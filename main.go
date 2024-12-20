package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	version = "1.0.0"
)

type Resource struct {
	Kind string
	Name string
	Children []*Resource
}

func main() {
	var kubeconfig *string
	var showVersion bool
	var namespace string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.StringVar(&namespace, "n", "", "namespace to show tree for (defaults to current namespace)")
	flag.Parse()

	if showVersion {
		fmt.Printf("kubectl-tree version %s\n", version)
		os.Exit(0)
	}

	// Create kubernetes client
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Get current namespace if not specified
	if namespace == "" {
		if ns, err := getCurrentNamespace(*kubeconfig); err == nil {
			namespace = ns
		} else {
			namespace = "default"
		}
	}

	// Build and display resource tree
	tree := buildResourceTree(clientset, namespace)
	printTree(tree, "", true)
}

func getCurrentNamespace(kubeconfig string) (string, error) {
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return "", err
	}

	context := config.Contexts[config.CurrentContext]
	if context != nil {
		return context.Namespace, nil
	}

	return "", fmt.Errorf("no current context found")
}

func buildResourceTree(clientset *kubernetes.Clientset, namespace string) *Resource {
	root := &Resource{
		Kind: "Namespace",
		Name: namespace,
		Children: make([]*Resource, 0),
	}

	// Get deployments
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, dep := range deployments.Items {
			depNode := &Resource{
				Kind: "Deployment",
				Name: dep.Name,
				Children: make([]*Resource, 0),
			}
			root.Children = append(root.Children, depNode)

			// Get ReplicaSets owned by this deployment
			rsList, err := clientset.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{})
			if err == nil {
				for _, rs := range rsList.Items {
					for _, owner := range rs.OwnerReferences {
						if owner.Kind == "Deployment" && owner.Name == dep.Name {
							rsNode := &Resource{
								Kind: "ReplicaSet",
								Name: rs.Name,
								Children: make([]*Resource, 0),
							}
							depNode.Children = append(depNode.Children, rsNode)

							// Get Pods owned by this ReplicaSet
							pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
							if err == nil {
								for _, pod := range pods.Items {
									for _, podOwner := range pod.OwnerReferences {
										if podOwner.Kind == "ReplicaSet" && podOwner.Name == rs.Name {
											podNode := &Resource{
												Kind: "Pod",
												Name: pod.Name,
												Children: make([]*Resource, 0),
											}
											rsNode.Children = append(rsNode.Children, podNode)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Get Services
	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, svc := range services.Items {
			svcNode := &Resource{
				Kind: "Service",
				Name: svc.Name,
				Children: make([]*Resource, 0),
			}
			root.Children = append(root.Children, svcNode)
		}
	}

	// Get ConfigMaps
	configMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, cm := range configMaps.Items {
			if len(cm.OwnerReferences) == 0 {
				cmNode := &Resource{
					Kind: "ConfigMap",
					Name: cm.Name,
					Children: make([]*Resource, 0),
				}
				root.Children = append(root.Children, cmNode)
			}
		}
	}

	// Get Secrets
	secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, secret := range secrets.Items {
			if len(secret.OwnerReferences) == 0 {
				secretNode := &Resource{
					Kind: "Secret",
					Name: secret.Name,
					Children: make([]*Resource, 0),
				}
				root.Children = append(root.Children, secretNode)
			}
		}
	}

	return root
}

func printTree(node *Resource, prefix string, isLast bool) {
	if node == nil {
		return
	}

	// Print current node
	fmt.Printf("%s%s%s/%s\n", prefix, getConnector(isLast), node.Kind, node.Name)

	// Prepare prefix for children
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	// Print children
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		printTree(child, childPrefix, isLastChild)
	}
}

func getConnector(isLast bool) string {
	if isLast {
		return "└── "
	}
	return "├── "
}