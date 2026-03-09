package deps

import (
	// Ensure these dependencies are pinned in go.mod even before they are used.
	_ "github.com/charmbracelet/lipgloss"
	_ "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/kubernetes"
)
