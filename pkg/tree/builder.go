package tree

import (
	"fmt"
	"kubectl-tree/pkg/k8s"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Builder handles building the resource tree
type Builder struct {
	client    *k8s.Client
	debug     bool
	resources *k8s.Resources // Add this field
}

// Update BuildTree to store resources
func (b *Builder) BuildTree(namespace string) (*Resource, error) {
	// Get all resources in namespace
	resources, err := b.client.GetResources(namespace)
	if err != nil {
		return nil, err
	}
	
	// Store resources for later use
	b.resources = resources
	
	// Check if namespace is empty
	if len(resources.Deployments.Items) == 0 &&
		len(resources.StatefulSets.Items) == 0 &&
		len(resources.DaemonSets.Items) == 0 &&
		len(resources.Jobs.Items) == 0 &&
		len(resources.CronJobs.Items) == 0 {
		fmt.Printf("No resources found in %s namespace.\n", namespace)
		return nil, nil
	}

	root := &Resource{
		Kind:     "Namespace",
		Name:     namespace,
		Children: make([]*Resource, 0),
	}

	// Create shared found map
	found := make(map[string]bool)

	// Add Deployments
	for i := range resources.Deployments.Items {
		dep := &resources.Deployments.Items[i]

		depNode := &Resource{
			Kind:     "Deployment",
			Name:     dep.Name,
			Children: make([]*Resource, 0),
		}
		root.Children = append(root.Children, depNode)

		// Add related resources
		b.addRelatedResources(dep, depNode, resources, found)

		// Add ReplicaSets
		for _, rs := range resources.GetReplicaSetsByOwner("Deployment", dep.Name) {
			rsNode := &Resource{
				Kind:     "ReplicaSet",
				Name:     rs.Name,
				Children: make([]*Resource, 0),
			}
			depNode.Children = append(depNode.Children, rsNode)

			// Add Pods
			for _, pod := range resources.GetPodsByOwner("ReplicaSet", rs.Name) {
			    podNode := &Resource{
			        Kind:     "Pod",
			        Name:     pod.Name,
			        Children: make([]*Resource, 0),
			    }
			    rsNode.Children = append(rsNode.Children, podNode)
			    
			    // Add containers to the pod
			    for _, container := range pod.Spec.Containers {
			        containerNode := &Resource{
			            Kind:     "Container",
			            Name:     container.Name,
			            Children: make([]*Resource, 0),
			        }
			        podNode.Children = append(podNode.Children, containerNode)
			    }
			}
			
			// Remove the call to addContainersToTree at the end of BuildTree
			// b.addContainersToTree(root)
		}
	}

	// Add StatefulSets
	for i := range resources.StatefulSets.Items {
		sts := &resources.StatefulSets.Items[i]
		stsNode := &Resource{
			Kind:     "StatefulSet",
			Name:     sts.Name,
			Children: make([]*Resource, 0),
		}
		root.Children = append(root.Children, stsNode)

		// Add related resources first
		b.addRelatedResources(sts, stsNode, resources, found)

		// Add Pods last so they appear after the related resources
		for _, pod := range resources.GetPodsByOwner("StatefulSet", sts.Name) {
			podNode := &Resource{
				Kind:     "Pod",
				Name:     pod.Name,
				Children: make([]*Resource, 0),
			}
			stsNode.Children = append(stsNode.Children, podNode)
			
			// Add containers to the pod
			for _, container := range pod.Spec.Containers {
				containerNode := &Resource{
					Kind:     "Container",
					Name:     container.Name,
					Children: make([]*Resource, 0),
				}
				podNode.Children = append(podNode.Children, containerNode)
			}
		}
	}

	// Add DaemonSets
	for i := range resources.DaemonSets.Items {
		ds := &resources.DaemonSets.Items[i]
		dsNode := &Resource{
			Kind:     "DaemonSet",
			Name:     ds.Name,
			Children: make([]*Resource, 0),
		}
		root.Children = append(root.Children, dsNode)

		// Add related resources
		b.addRelatedResources(ds, dsNode, resources, found)

		// Add Pods
		for _, pod := range resources.GetPodsByOwner("DaemonSet", ds.Name) {
			podNode := &Resource{
				Kind:     "Pod",
				Name:     pod.Name,
				Children: make([]*Resource, 0),
			}
			dsNode.Children = append(dsNode.Children, podNode)
			
			// Add containers to the pod
			for _, container := range pod.Spec.Containers {
				containerNode := &Resource{
					Kind:     "Container",
					Name:     container.Name,
					Children: make([]*Resource, 0),
				}
				podNode.Children = append(podNode.Children, containerNode)
			}
		}
	}

	// Add standalone Jobs (not owned by CronJobs)
	for i := range resources.Jobs.Items {
		job := &resources.Jobs.Items[i]
		if len(job.OwnerReferences) == 0 || job.OwnerReferences[0].Kind != "CronJob" {
			jobNode := &Resource{
				Kind:     "Job",
				Name:     job.Name,
				Children: make([]*Resource, 0),
			}
			root.Children = append(root.Children, jobNode)

			// Add related resources
			b.addRelatedResources(job, jobNode, resources, found)

			// Add Pods
			for _, pod := range resources.GetPodsByOwner("Job", job.Name) {
				podNode := &Resource{
					Kind:     "Pod",
					Name:     pod.Name,
					Children: make([]*Resource, 0),
				}
				jobNode.Children = append(jobNode.Children, podNode)
				
				// Add containers to the pod
				for _, container := range pod.Spec.Containers {
					containerNode := &Resource{
						Kind:     "Container",
						Name:     container.Name,
						Children: make([]*Resource, 0),
					}
					podNode.Children = append(podNode.Children, containerNode)
				}
			}
		}
	}

	// Add CronJobs
	for i := range resources.CronJobs.Items {
		cronJob := &resources.CronJobs.Items[i]
		cronJobNode := &Resource{
			Kind:     "CronJob",
			Name:     cronJob.Name,
			Children: make([]*Resource, 0),
		}
		root.Children = append(root.Children, cronJobNode)

		// Add Jobs owned by this CronJob
		for _, job := range resources.GetJobsByOwner("CronJob", cronJob.Name) {
			jobNode := &Resource{
				Kind:     "Job",
				Name:     job.Name,
				Children: make([]*Resource, 0),
			}
			cronJobNode.Children = append(cronJobNode.Children, jobNode)

			// Add Pods
			for _, pod := range resources.GetPodsByOwner("Job", job.Name) {
				podNode := &Resource{
					Kind:     "Pod",
					Name:     pod.Name,
					Children: make([]*Resource, 0),
				}
				jobNode.Children = append(jobNode.Children, podNode)
				
				// Add containers to the pod
				for _, container := range pod.Spec.Containers {
					containerNode := &Resource{
						Kind:     "Container",
						Name:     container.Name,
						Children: make([]*Resource, 0),
					}
					podNode.Children = append(podNode.Children, containerNode)
				}
			}
		}
	}

	// Remove this line to prevent adding containers twice
	// b.addContainersToTree(root)

	return root, nil
}

// New method to add containers to pods in the tree
// Fix the addContainersToTree method to use Resource instead of node
// Fix the addContainersToTree method to properly handle errors and use the correct parameters
func (b *Builder) addContainersToTree(node *Resource) {
    // Skip if node is nil
    if node == nil {
        return
    }
    
    // Process Pod nodes
    if node.Kind == "Pod" {
        // Find the pod in the already fetched resources instead of making a new API call
        for _, child := range node.Children {
            // Skip if already processed
            if child.Kind == "Container" {
                continue
            }
        }
        
        // Just add containers from the pod spec directly
        // This is safer than making additional API calls
        for _, pod := range b.resources.Pods.Items {
            if pod.Name == node.Name {
                for _, container := range pod.Spec.Containers {
                    containerNode := &Resource{
                        Kind:     "Container",
                        Name:     container.Name,
                        Children: make([]*Resource, 0),
                    }
                    node.Children = append(node.Children, containerNode)
                }
                break
            }
        }
    }
    
    // Process children recursively
    for _, child := range node.Children {
        b.addContainersToTree(child)
    }
}

// addRelatedResources adds related resources as children of the workload node
func (b *Builder) addRelatedResources(workload metav1.Object, workloadNode *Resource, resources *k8s.Resources, found map[string]bool) {
	// Get the PodSpec from the workload
	var podSpec *corev1.PodSpec
	switch w := workload.(type) {
	case *appsv1.StatefulSet:
		podSpec = &w.Spec.Template.Spec
	case *appsv1.Deployment:
		podSpec = &w.Spec.Template.Spec
	case *appsv1.DaemonSet:
		podSpec = &w.Spec.Template.Spec
	case *batchv1.Job:
		podSpec = &w.Spec.Template.Spec
	case *batchv1.CronJob:
		podSpec = &w.Spec.JobTemplate.Spec.Template.Spec
	default:
		return // Early return if workload type is not supported
	}

	// Find related resources
	services, configMaps, secrets, pvcs := resources.FindRelatedResources(workload, podSpec, found, b.debug)

	if b.debug {
		fmt.Printf("Debug: Found resources for %s: secrets=%d, pvcs=%d, configmaps=%d, services=%d\n",
			workload.GetName(), len(secrets), len(pvcs), len(configMaps), len(services))
	}

	// Add Services
	for _, svc := range services {
		if b.debug {
			fmt.Printf("\tDebug: Adding Service %s to %s\n", svc.Name, workload.GetName())
		}
		svcNode := &Resource{
			Kind: "Service",
			Name: svc.Name,

			Children: make([]*Resource, 0),
		}
		workloadNode.Children = append(workloadNode.Children, svcNode)
	}

	// Add ConfigMaps
	for _, cm := range configMaps {
		if b.debug {
			fmt.Printf("\tDebug: Adding ConfigMap %s to %s\n", cm.Name, workload.GetName())
		}
		cmNode := &Resource{
			Kind:     "ConfigMap",
			Name:     cm.Name,
			Children: make([]*Resource, 0),
		}
		workloadNode.Children = append(workloadNode.Children, cmNode)
	}

	// Add Secrets
	for _, secret := range secrets {
		if b.debug {
			fmt.Printf("\tDebug: Adding Secret %s to %s\n", secret.Name, workload.GetName())
		}
		secretNode := &Resource{
			Kind:     "Secret",
			Name:     secret.Name,
			Children: make([]*Resource, 0),
		}
		workloadNode.Children = append(workloadNode.Children, secretNode)
	}

	// Add PVCs
	for _, pvc := range pvcs {
		if b.debug {
			fmt.Printf("\tDebug: Adding PVC %s to %s\n", pvc.Name, workload.GetName())
		}
		pvcNode := &Resource{
			Kind:     "PersistentVolumeClaim",
			Name:     pvc.Name,
			Children: make([]*Resource, 0),
		}
		workloadNode.Children = append(workloadNode.Children, pvcNode)
	}
}

// NewBuilder creates a new tree builder
func NewBuilder(client *k8s.Client, debug bool) *Builder {
	return &Builder{
		client:    client,
		debug:     debug,
		resources: nil, // Will be set in BuildTree
	}
}
