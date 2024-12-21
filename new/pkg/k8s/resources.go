package k8s

import (
	"strings"

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

	// Find related Services
	for i, svc := range r.Services.Items {
		if svc.Spec.Selector == nil {
			continue
		}

		matches := true
		for key, value := range svc.Spec.Selector {
			if workloadLabels[key] != value {
				matches = false
				break
			}
		}
		
		if !matches && (strings.HasPrefix(svc.Name, workloadName) || 
					  strings.HasSuffix(svc.Name, workloadName)) {
			matches = true
		}

		if matches {
			key := "Service/" + svc.Name
			if !found[key] {
				services = append(services, &r.Services.Items[i])
				found[key] = true
			}
		}
	}

	// Find related resources from volumes
	for _, vol := range podSpec.Volumes {
		if vol.ConfigMap != nil {
			key := "ConfigMap/" + vol.ConfigMap.Name
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
			key := "Secret/" + vol.Secret.SecretName
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
			key := "PVC/" + vol.PersistentVolumeClaim.ClaimName
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

	return services, configMaps, secrets, pvcs
}