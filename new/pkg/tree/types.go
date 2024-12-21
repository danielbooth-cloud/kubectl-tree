package tree

import (
	"sort"
)

// Resource represents a Kubernetes resource in the tree
type Resource struct {
	Kind     string
	Name     string
	Children []*Resource
}

// NewResource creates a new resource node
func NewResource(kind, name string) *Resource {
	return &Resource{
		Kind:     kind,
		Name:     name,
		Children: make([]*Resource, 0),
	}
}

// AddChild adds a child resource to this resource
func (r *Resource) AddChild(child *Resource) {
	r.Children = append(r.Children, child)
}

// SortChildren sorts child resources by Kind and Name
func (r *Resource) SortChildren() {
	// Sort children recursively
	sort.Slice(r.Children, func(i, j int) bool {
		// First sort by Kind
		if r.Children[i].Kind != r.Children[j].Kind {
			return r.Children[i].Kind < r.Children[j].Kind
		}
		// Then by Name
		return r.Children[i].Name < r.Children[j].Name
	})

	// Sort children of children
	for _, child := range r.Children {
		child.SortChildren()
	}
}

// ResourceKind represents the type of Kubernetes resource
type ResourceKind string

// Define constants for resource kinds to avoid typos
const (
	KindNamespace             ResourceKind = "Namespace"
	KindDeployment           ResourceKind = "Deployment"
	KindStatefulSet          ResourceKind = "StatefulSet"
	KindDaemonSet            ResourceKind = "DaemonSet"
	KindReplicaSet           ResourceKind = "ReplicaSet"
	KindPod                  ResourceKind = "Pod"
	KindService              ResourceKind = "Service"
	KindConfigMap            ResourceKind = "ConfigMap"
	KindSecret               ResourceKind = "Secret"
	KindPersistentVolumeClaim ResourceKind = "PersistentVolumeClaim"
	KindJob                  ResourceKind = "Job"
	KindCronJob              ResourceKind = "CronJob"
)

// IsWorkload returns true if the resource kind is a workload
func (k ResourceKind) IsWorkload() bool {
	switch k {
	case KindDeployment, KindStatefulSet, KindDaemonSet, KindJob, KindCronJob:
		return true
	default:
		return false
	}
}

// IsPod returns true if the resource kind is a Pod
func (k ResourceKind) IsPod() bool {
	return k == KindPod
}

// IsController returns true if the resource kind is a controller
func (k ResourceKind) IsController() bool {
	switch k {
	case KindDeployment, KindStatefulSet, KindDaemonSet, KindReplicaSet, KindJob, KindCronJob:
		return true
	default:
		return false
	}
}

// String returns the string representation of the ResourceKind
func (k ResourceKind) String() string {
	return string(k)
}

// ResourceOrder defines the order in which resources should be displayed
var ResourceOrder = map[ResourceKind]int{
	KindDeployment:           1,
	KindStatefulSet:          2,
	KindDaemonSet:            3,
	KindJob:                  4,
	KindCronJob:             5,
	KindReplicaSet:          6,
	KindPod:                 7,
	KindService:             8,
	KindConfigMap:           9,
	KindSecret:              10,
	KindPersistentVolumeClaim: 11,
}

// GetResourceOrder returns the display order for a resource kind
func GetResourceOrder(kind string) int {
	if order, exists := ResourceOrder[ResourceKind(kind)]; exists {
		return order
	}
	return 100 // Put unknown types at the end
}