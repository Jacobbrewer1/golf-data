package utils

import "os"

const (
	// kubeNamespacePath is the path to the Kubernetes namespace file
	kubernetesNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var PodName = os.Getenv("HOSTNAME")

func GetDeployedKubernetesNamespace() (string, error) {
	got, err := os.ReadFile(kubernetesNamespacePath)
	if err != nil {
		return "", err
	}
	return string(got), nil
}
