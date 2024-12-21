package tree

// Resource represents a Kubernetes resource in the tree
type Resource struct {
	Kind     string
	Name     string
	Children []*Resource
}
