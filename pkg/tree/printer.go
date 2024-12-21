package tree

import (
	"fmt"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

type Printer struct {
	useColor bool
}

func NewPrinter(useColor bool) *Printer {
	return &Printer{
		useColor: useColor,
	}
}

func (p *Printer) getResourceColor(kind string) string {
	if !p.useColor {
		return ""
	}

	switch kind {
	case "Deployment", "StatefulSet", "DaemonSet":
		return colorBlue
	case "Pod":
		return colorGreen
	case "Service":
		return colorYellow
	case "ConfigMap", "Secret":
		return colorPurple
	case "PersistentVolumeClaim":
		return colorCyan
	default:
		return ""
	}
}

func (p *Printer) getConnector(isLast bool) string {
	if isLast {
		return "└── "
	}
	return "├── "
}

func (p *Printer) PrintTree(node *Resource, prefix string, isLast bool) {
	if node == nil {
		return
	}

	color := p.getResourceColor(node.Kind)
	fmt.Printf("%s%s%s%s/%s%s\n",
		prefix,
		p.getConnector(isLast),
		color,
		node.Kind,
		node.Name,
		colorReset,
	)

	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		p.PrintTree(child, childPrefix, isLastChild)
	}
}
