package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes clientset
type Client struct {
	clientset *kubernetes.Clientset
}

// Resources holds all the resources fetched from the cluster
type Resources struct {
	Services     *corev1.ServiceList
	ConfigMaps   *corev1.ConfigMapList
	Secrets      *corev1.SecretList
	PVCs         *corev1.PersistentVolumeClaimList
	Pods         *corev1.PodList
	Deployments  *appsv1.DeploymentList
	StatefulSets *appsv1.StatefulSetList
	DaemonSets   *appsv1.DaemonSetList
	ReplicaSets  *appsv1.ReplicaSetList
	Jobs         *batchv1.JobList
	CronJobs     *batchv1.CronJobList
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfig string) (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client: %v", err)
	}

	return &Client{clientset: clientset}, nil
}

// NamespaceExists checks if a namespace exists
func (c *Client) NamespaceExists(namespace string) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	return err
}

// GetResources fetches all resources from the specified namespace
func (c *Client) GetResources(namespace string) (*Resources, error) {
	ctx := context.TODO()
	opts := metav1.ListOptions{}

	resources := &Resources{}
	var err error

	// Fetch Services
	resources.Services, err = c.clientset.CoreV1().Services(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching services: %v", err)
	}

	// Fetch ConfigMaps
	resources.ConfigMaps, err = c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching configmaps: %v", err)
	}

	// Fetch Secrets
	resources.Secrets, err = c.clientset.CoreV1().Secrets(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching secrets: %v", err)
	}

	// Fetch PVCs
	resources.PVCs, err = c.clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching pvcs: %v", err)
	}

	// Fetch Pods
	resources.Pods, err = c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching pods: %v", err)
	}

	// Fetch Deployments
	resources.Deployments, err = c.clientset.AppsV1().Deployments(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching deployments: %v", err)
	}

	// Fetch StatefulSets
	resources.StatefulSets, err = c.clientset.AppsV1().StatefulSets(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching statefulsets: %v", err)
	}

	// Fetch DaemonSets
	resources.DaemonSets, err = c.clientset.AppsV1().DaemonSets(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching daemonsets: %v", err)
	}

	// Fetch ReplicaSets
	resources.ReplicaSets, err = c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching replicasets: %v", err)
	}

	// Fetch Jobs
	resources.Jobs, err = c.clientset.BatchV1().Jobs(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching jobs: %v", err)
	}

	// Fetch CronJobs
	resources.CronJobs, err = c.clientset.BatchV1().CronJobs(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching cronjobs: %v", err)
	}

	return resources, nil
}