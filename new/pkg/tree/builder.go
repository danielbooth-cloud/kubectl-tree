package tree

import (
	"kubectl-tree/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Builder handles building the resource tree
type Builder struct {
	client *k8s.Client
}

// NewBuilder creates a new tree builder
func NewBuilder(client *k8s.Client) *Builder {
	return &Builder{client: client}
}

// BuildTree builds a tree of resources in the specified namespace
func (b *Builder) BuildTree(namespace string) (*Resource, error) {
	// Get all resources
	resources, err := b.client.GetResources(namespace)
	if err != nil {
		return nil, err
	}

	root := &Resource{
		Kind: "Namespace",
		Name: namespace,
		Children: make([]*Resource, 0),
	}

	// Add Deployments
	for i := range resources.Deployments.Items {
		dep := &resources.Deployments.Items[i]
		depNode := &Resource{
			Kind: "Deployment",
			Name: dep.Name,
			Children: make([]*Resource, 0),
		}
		root.Children = append(root.Children, depNode)

		// Add related resources
		b.addRelatedResources(dep, depNode, resources)

		// Add ReplicaSets
		for _, rs := range resources.GetReplicaSetsByOwner("Deployment", dep.Name) {
			rsNode := &Resource{
				Kind: "ReplicaSet",
				Name: rs.Name,
				Children: make([]*Resource, 0),
			}
			depNode.Children = append(depNode.Children, rsNode)

			// Add Pods
			for _, pod := range resources.GetPodsByOwner("ReplicaSet", rs.Name) {
				podNode := &Resource{
					Kind: "Pod",
					Name: pod.Name,
					Children: make([]*Resource, 0),
				}
				rsNode.Children = append(rsNode.Children, podNode)
			}
		}
	}

	// Add StatefulSets
	for i := range resources.StatefulSets.Items {
		sts := &resources.StatefulSets.Items[i]
		stsNode := &Resource{
			Kind: "StatefulSet",
			Name: sts.Name,
			Children: make([]*Resource, 0),
		}
		root.Children = append(root.Children, stsNode)

		// Add related resources
		b.addRelatedResources(sts, stsNode, resources)

		// Add Pods
		for _, pod := range resources.GetPodsByOwner("StatefulSet", sts.Name) {
			podNode := &Resource{
				Kind: "Pod",
				Name: pod.Name,
				Children: make([]*Resource, 0),
			}
			stsNode.Children = append(stsNode.Children, podNode)
		}
	}

	// Add DaemonSets
	for i := range resources.DaemonSets.Items {
		ds := &resources.DaemonSets.Items[i]
		dsNode := &Resource{
			Kind: "DaemonSet",
			Name: ds.Name,
			Children: make([]*Resource, 0),
		}
		root.Children = append(root.Children, dsNode)

		// Add related resources
		b.addRelatedResources(ds, dsNode, resources)

		// Add Pods
		for _, pod := range resources.GetPodsByOwner("DaemonSet", ds.Name) {
			podNode := &Resource{
				Kind: "Pod",
				Name: pod.Name,
				Children: make([]*Resource, 0),
			}
			dsNode.Children = append(dsNode.Children, podNode)
		}
	}

	// Add standalone Jobs (not owned by CronJobs)
	for i := range resources.Jobs.Items {
		job := &resources.Jobs.Items[i]
		if len(job.OwnerReferences) == 0 || job.OwnerReferences[0].Kind != "CronJob" {
			jobNode := &Resource{
				Kind: "Job",
				Name: job.Name,
				Children: make([]*Resource, 0),
			}
			root.Children = append(root.Children, jobNode)

			// Add related resources
			b.addRelatedResources(job, jobNode, resources)

			// Add Pods
			for _, pod := range resources.GetPodsByOwner("Job", job.Name) {
				podNode := &Resource{
					Kind: "Pod",
					Name: pod.Name,
					Children: make([]*Resource, 0),
				}
				jobNode.Children = append(jobNode.Children, podNode)
			}
		}
	}

	// Add CronJobs
	for i := range resources.CronJobs.Items {
		cronJob := &resources.CronJobs.Items[i]
		cronJobNode := &Resource{
			Kind: "CronJob",
			Name: cronJob.Name,
			Children: make([]*Resource, 0),
		}
		root.Children = append(root.Children, cronJobNode)

		// Add Jobs owned by this CronJob
		for _, job := range resources.GetJobsByOwner("CronJob", cronJob.Name) {
			jobNode := &Resource{
				Kind: "Job",
				Name: job.Name,
				Children: make([]*Resource, 0),
			}
			cronJobNode.Children = append(cronJobNode.Children, jobNode)

			// Add Pods
			for _, pod := range resources.GetPodsByOwner("Job", job.Name) {
				podNode := &Resource{
					Kind: "Pod",
					Name: pod.Name,
					Children: make([]*Resource, 0),
				}
				jobNode.Children = append(jobNode.Children, podNode)
			}
		}
	}

	return root, nil
}

// addRelatedResources adds related resources as children of the workload node
func (b *Builder) addRelatedResources(workload metav1.Object, workloadNode *Resource, resources *k8s.Resources) {
	services, configMaps, secrets, pvcs := resources.FindRelatedResources(workload)

	// Add Services
	for _, svc := range services {
		svcNode := &Resource{
			Kind: "Service",
			Name: svc.Name,
			Children: make([]*Resource, 0),
		}
		workloadNode.Children = append(workloadNode.Children, svcNode)
	}

	// Add ConfigMaps
	for _, cm := range configMaps {
		cmNode := &Resource{
			Kind: "ConfigMap",
			Name: cm.Name,
			Children: make([]*Resource, 0),
		}
		workloadNode.Children = append(workloadNode.Children, cmNode)
	}

	// Add Secrets
	for _, secret := range secrets {
		secretNode := &Resource{
			Kind: "Secret",
			Name: secret.Name,
			Children: make([]*Resource, 0),
		}
		workloadNode.Children = append(workloadNode.Children, secretNode)
	}

	// Add PVCs
	for _, pvc := range pvcs {
		pvcNode := &Resource{
			Kind: "PersistentVolumeClaim",
			Name: pvc.Name,
			Children: make([]*Resource, 0),
		}
		workloadNode.Children = append(workloadNode.Children, pvcNode)
	}
}