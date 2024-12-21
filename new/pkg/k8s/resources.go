package k8s

import (
	"fmt"
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
func (r *Resources) FindRelatedResources(workload metav1.Object, podSpec *corev1.PodSpec, found map[string]bool, debug bool) ([]*corev1.Service, []*corev1.ConfigMap, []*corev1.Secret, []*corev1.PersistentVolumeClaim) {
	var (
		services   []*corev1.Service
		configMaps []*corev1.ConfigMap
		secrets    []*corev1.Secret
		pvcs       []*corev1.PersistentVolumeClaim
	)

	workloadLabels := workload.GetLabels()
	workloadName := workload.GetName()
	
	// Get workload kind and check for StatefulSet VolumeClaimTemplates
	var workloadKind string
	switch w := workload.(type) {
	case *appsv1.StatefulSet:
		workloadKind = "StatefulSet"
		// Check VolumeClaimTemplates for StatefulSets
		for _, template := range w.Spec.VolumeClaimTemplates {
			pvcName := template.Name + "-" + workloadName + "-0"
			if debug {
				fmt.Printf("Debug: Looking for StatefulSet VolumeClaimTemplate PVC: %s\n", pvcName)
			}
			for i, pvc := range r.PVCs.Items {
				if pvc.Name == pvcName {
					if debug {
						fmt.Printf("Debug: Found StatefulSet VolumeClaimTemplate PVC: %s\n", pvc.Name)
					}
					pvcs = append(pvcs, &r.PVCs.Items[i])
				}
			}
		}
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
		if debug {
			fmt.Printf("Debug: Checking volumes for %s (%s): count=%d\n", 
				workloadName, workloadKind, len(podSpec.Volumes))
			// Print all PVCs in namespace for debugging
			fmt.Printf("Debug: Available PVCs in namespace:\n")
			for _, pvc := range r.PVCs.Items {
				fmt.Printf("  - %s\n", pvc.Name)
			}
		}

		// Check volumes
		for _, vol := range podSpec.Volumes {
			if vol.ConfigMap != nil {
				for i, cm := range r.ConfigMaps.Items {
					if cm.Name == vol.ConfigMap.Name {
						configMaps = append(configMaps, &r.ConfigMaps.Items[i])
						break
					}
				}
			}

			if vol.Secret != nil {
				for i, secret := range r.Secrets.Items {
					if secret.Name == vol.Secret.SecretName {
						secrets = append(secrets, &r.Secrets.Items[i])
						break
					}
				}
			}

			if vol.PersistentVolumeClaim != nil {
				if debug {
					fmt.Printf("Debug: Checking PVC %s for %s\n", 
						vol.PersistentVolumeClaim.ClaimName, workloadName)
				}
				pvcFound := false
				for i, pvc := range r.PVCs.Items {
					// Direct name match
					if pvc.Name == vol.PersistentVolumeClaim.ClaimName {
						if debug {
							fmt.Printf("Debug: Found PVC by direct match: %s\n", pvc.Name)
						}
						pvcs = append(pvcs, &r.PVCs.Items[i])
						pvcFound = true
						break
					}
				}

				// For StatefulSets, also check for PVCs that match the pattern
				if !pvcFound && workloadKind == "StatefulSet" {
					for i, pvc := range r.PVCs.Items {
						// Check for common StatefulSet PVC patterns
						if strings.Contains(pvc.Name, workloadName) {
							if debug {
								fmt.Printf("Debug: Found StatefulSet PVC: %s for %s\n", 
									pvc.Name, workloadName)
							}
							pvcs = append(pvcs, &r.PVCs.Items[i])
						}
					}
				}
			}
		}

		// Check environment variables
		for _, container := range podSpec.Containers {
			// Check envFrom
			for _, envFrom := range container.EnvFrom {
				if envFrom.SecretRef != nil {
					for i, secret := range r.Secrets.Items {
						if secret.Name == envFrom.SecretRef.Name {
							if debug {
								fmt.Printf("Debug: Found Secret from envFrom: %s\n", secret.Name)
							}
							secrets = append(secrets, &r.Secrets.Items[i])
							break
						}
					}
				}
			}
			
			// Check individual env variables
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
					for i, secret := range r.Secrets.Items {
						if secret.Name == env.ValueFrom.SecretKeyRef.Name {
							if debug {
								fmt.Printf("Debug: Found Secret from env: %s\n", secret.Name)
							}
							secrets = append(secrets, &r.Secrets.Items[i])
							break
						}
					}
				}
			}
		}
	}

	return services, configMaps, secrets, pvcs
}