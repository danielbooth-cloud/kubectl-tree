package k8s

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPodsByOwner returns all pods owned by the specified owner
func (r *Resources) GetPodsByOwner(ownerKind, ownerName string) []*corev1.Pod {
	var pods []*corev1.Pod
	for i, pod := range r.Pods.Items {
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == ownerKind && owner.Name == ownerName {
				pods = append(pods, &r.Pods.Items[i])
				break
			}
		}
	}
	return pods
}

// GetReplicaSetsByOwner returns all ReplicaSets owned by the specified owner
func (r *Resources) GetReplicaSetsByOwner(ownerKind, ownerName string) []*appsv1.ReplicaSet {
	var replicaSets []*appsv1.ReplicaSet
	for i, rs := range r.ReplicaSets.Items {
		for _, owner := range rs.OwnerReferences {
			if owner.Kind == ownerKind && owner.Name == ownerName {
				replicaSets = append(replicaSets, &r.ReplicaSets.Items[i])
				break
			}
		}
	}
	return replicaSets
}

// GetJobsByOwner returns all Jobs owned by the specified owner
func (r *Resources) GetJobsByOwner(ownerKind, ownerName string) []*batchv1.Job {
	var jobs []*batchv1.Job
	for i, job := range r.Jobs.Items {
		for _, owner := range job.OwnerReferences {
			if owner.Kind == ownerKind && owner.Name == ownerName {
				jobs = append(jobs, &r.Jobs.Items[i])
				break
			}
		}
	}
	return jobs
}

// FindRelatedResources finds all resources related to a workload
func (r *Resources) FindRelatedResources(workload metav1.Object, podSpec *corev1.PodSpec, found map[string]bool) ([]*corev1.Service, []*corev1.ConfigMap, []*corev1.Secret, []*corev1.PersistentVolumeClaim) {
	var (
		services   []*corev1.Service
		configMaps []*corev1.ConfigMap
		secrets    []*corev1.Secret
		pvcs       []*corev1.PersistentVolumeClaim
	)

	workloadLabels := workload.GetLabels()
	workloadName := workload.GetName()
	
	// Get workload kind
	var workloadKind string
	switch workload.(type) {
	case *appsv1.StatefulSet:
		workloadKind = "StatefulSet"
	case *appsv1.Deployment:
		workloadKind = "Deployment"
	case *appsv1.DaemonSet:
		workloadKind = "DaemonSet"
	case *batchv1.Job:
		workloadKind = "Job"
	case *batchv1.CronJob:
		workloadKind = "CronJob"
	default:
		workloadKind = "Unknown"
	}

	// Find related Services - don't use found map for services since they can be shared
	for i, svc := range r.Services.Items {
		if svc.Spec.Selector == nil {
			continue
		}

		matches := false
		// Check if service selector matches workload labels
		if len(svc.Spec.Selector) > 0 {
			matches = true
			for key, value := range svc.Spec.Selector {
				if workloadLabels[key] != value {
					matches = false
					break
				}
			}
		}

		// For StatefulSets, also check if service name matches workload name
		if !matches && workloadKind == "StatefulSet" {
			if svc.Name == workloadName || 
			   svc.Name == workloadName+"-headless" {
				matches = true
			}
		}

		if matches {
			services = append(services, &r.Services.Items[i])
		}
	}

	// Find related resources from volumes and environment
	if podSpec != nil {
		// Check volumes
		for _, vol := range podSpec.Volumes {
			if vol.ConfigMap != nil {
				key := fmt.Sprintf("%s/ConfigMap/%s", workloadKind, vol.ConfigMap.Name)
				if !found[key] {
					for i, cm := range r.ConfigMaps.Items {
						if cm.Name == vol.ConfigMap.Name {
							configMaps = append(configMaps, &r.ConfigMaps.Items[i])
							found[key] = true
							break
						}
					}
				}
			}

			if vol.Secret != nil {
				key := fmt.Sprintf("%s/Secret/%s", workloadKind, vol.Secret.SecretName)
				if !found[key] {
					for i, secret := range r.Secrets.Items {
						if secret.Name == vol.Secret.SecretName {
							secrets = append(secrets, &r.Secrets.Items[i])
							found[key] = true
							break
						}
					}
				}
			}

			if vol.PersistentVolumeClaim != nil {
				key := fmt.Sprintf("%s/PVC/%s", workloadKind, vol.PersistentVolumeClaim.ClaimName)
				if !found[key] {
					for i, pvc := range r.PVCs.Items {
						if pvc.Name == vol.PersistentVolumeClaim.ClaimName {
							pvcs = append(pvcs, &r.PVCs.Items[i])
							found[key] = true
							break
						}
					}
				}
			}
		}

		// Check environment variables
		for _, container := range podSpec.Containers {
			for _, env := range container.EnvFrom {
				if env.ConfigMapRef != nil {
					key := fmt.Sprintf("%s/ConfigMap/%s", workloadKind, env.ConfigMapRef.Name)
					if !found[key] {
						for i, cm := range r.ConfigMaps.Items {
							if cm.Name == env.ConfigMapRef.Name {
								configMaps = append(configMaps, &r.ConfigMaps.Items[i])
								found[key] = true
								break
							}
						}
					}
				}
				if env.SecretRef != nil {
					key := fmt.Sprintf("%s/Secret/%s", workloadKind, env.SecretRef.Name)
					if !found[key] {
						for i, secret := range r.Secrets.Items {
							if secret.Name == env.SecretRef.Name {
								secrets = append(secrets, &r.Secrets.Items[i])
								found[key] = true
								break
							}
						}
					}
				}
			}
		}
	}

	return services, configMaps, secrets, pvcs
}