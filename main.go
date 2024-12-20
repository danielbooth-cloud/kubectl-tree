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

	// Get StatefulSets
	statefulSets, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, sts := range statefulSets.Items {
			stsNode := &Resource{
				Kind: "StatefulSet",
				Name: sts.Name,
				Children: make([]*Resource, 0),
			}
			root.Children = append(root.Children, stsNode)

			// Get Pods owned by this StatefulSet
			pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
			if err == nil {
				for _, pod := range pods.Items {
					for _, podOwner := range pod.OwnerReferences {
						if podOwner.Kind == "StatefulSet" && podOwner.Name == sts.Name {
							podNode := &Resource{
								Kind: "Pod",
								Name: pod.Name,
								Children: make([]*Resource, 0),
							}
							stsNode.Children = append(stsNode.Children, podNode)
						}
					}
				}
			}
		}
	}

	// Get DaemonSets
	daemonSets, err := clientset.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, ds := range daemonSets.Items {
			dsNode := &Resource{
				Kind: "DaemonSet",
				Name: ds.Name,
				Children: make([]*Resource, 0),
			}
			root.Children = append(root.Children, dsNode)

			// Get Pods owned by this DaemonSet
			pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
			if err == nil {
				for _, pod := range pods.Items {
					for _, podOwner := range pod.OwnerReferences {
						if podOwner.Kind == "DaemonSet" && podOwner.Name == ds.Name {
							podNode := &Resource{
								Kind: "Pod",
								Name: pod.Name,
								Children: make([]*Resource, 0),
							}
							dsNode.Children = append(dsNode.Children, podNode)
						}
					}
				}
			}
		}
	}

	// Get Jobs
	jobs, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, job := range jobs.Items {
			// Skip jobs owned by CronJobs (they'll be added as children of CronJobs)
			if len(job.OwnerReferences) > 0 && job.OwnerReferences[0].Kind == "CronJob" {
				continue
			}
			
			jobNode := &Resource{
				Kind: "Job",
				Name: job.Name,
				Children: make([]*Resource, 0),
			}
			root.Children = append(root.Children, jobNode)

			// Get Pods owned by this Job
			pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
			if err == nil {
				for _, pod := range pods.Items {
					for _, podOwner := range pod.OwnerReferences {
						if podOwner.Kind == "Job" && podOwner.Name == job.Name {
							podNode := &Resource{
								Kind: "Pod",
								Name: pod.Name,
								Children: make([]*Resource, 0),
							}
							jobNode.Children = append(jobNode.Children, podNode)
						}
					}
				}
			}
		}
	}

	// Get CronJobs
	cronJobs, err := clientset.BatchV1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, cronJob := range cronJobs.Items {
			cronJobNode := &Resource{
				Kind: "CronJob",
				Name: cronJob.Name,
				Children: make([]*Resource, 0),
			}
			root.Children = append(root.Children, cronJobNode)

			// Get Jobs owned by this CronJob
			jobs, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
			if err == nil {
				for _, job := range jobs.Items {
					for _, jobOwner := range job.OwnerReferences {
						if jobOwner.Kind == "CronJob" && jobOwner.Name == cronJob.Name {
							jobNode := &Resource{
								Kind: "Job",
								Name: job.Name,
								Children: make([]*Resource, 0),
							}
							cronJobNode.Children = append(cronJobNode.Children, jobNode)

							// Get Pods owned by this Job
							pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
							if err == nil {
								for _, pod := range pods.Items {
									for _, podOwner := range pod.OwnerReferences {
										if podOwner.Kind == "Job" && podOwner.Name == job.Name {
											podNode := &Resource{
												Kind: "Pod",
												Name: pod.Name,
												Children: make([]*Resource, 0),
											}
											jobNode.Children = append(jobNode.Children, podNode)
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

	// Get PersistentVolumeClaims
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, pvc := range pvcs.Items {
			if len(pvc.OwnerReferences) == 0 {
				pvcNode := &Resource{
					Kind: "PersistentVolumeClaim",
					Name: pvc.Name,
					Children: make([]*Resource, 0),
				}
				root.Children = append(root.Children, pvcNode)
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