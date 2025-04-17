package utils

import (
	"os"
	"sync"
)

const (
	// kubeNamespacePath is the path to the Kubernetes namespace file
	kubernetesNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

// PodName returns the name of the pod. By default, Kubernetes sets the pod name as the HOSTNAME environment variable.
var PodName = sync.OnceValue(func() string {
	return os.Getenv("HOSTNAME")
})

// DeployedNamespace returns the namespace in which the pod is deployed. This is read from the Kubernetes namespace file.
var DeployedNamespace = sync.OnceValue(func() string {
	got, err := os.ReadFile(kubernetesNamespacePath)
	if err != nil {
		return "default" // Fallback to default namespace if the file is not found
	}
	return string(got)
})
