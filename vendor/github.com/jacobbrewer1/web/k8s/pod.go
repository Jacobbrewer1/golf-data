package k8s

import (
	"net"
	"os"
	"strings"
	"sync"
)

const (
	// kubernetesServiceAccountPath is the path to the Kubernetes service account file
	kubernetesServiceAccountPath = "/var/run/secrets/kubernetes.io/serviceaccount"

	// kubernetesNamespacePath is the path to the Kubernetes namespace file
	kubernetesNamespacePath = kubernetesServiceAccountPath + "/namespace"

	// kubernetesServiceAccountTokenPath is the path to the Kubernetes service account token file
	kubernetesServiceAccountTokenPath = kubernetesServiceAccountPath + "/token"
)

const (
	envHostname           = "HOSTNAME"
	envPodIP              = "POD_IP"
	envServiceAccountName = "SERVICE_ACCOUNT_NAME"
	envKubernetesHost     = "KUBERNETES_SERVICE_HOST"
	envKubernetesPort     = "KUBERNETES_SERVICE_PORT"
	envNodeName           = "NODE_NAME"
)

var (
	// PodName returns the name of the pod. By default, Kubernetes sets the pod name as the HOSTNAME environment variable.
	PodName = sync.OnceValue(func() string {
		hostname, err := os.Hostname()
		if err == nil {
			return hostname
		}

		// Attempt to read the pod name from the environment variable
		hostname = os.Getenv(envHostname)
		if hostname != "" {
			return hostname
		}

		// Fallback to an empty string if both methods fail
		return ""
	})

	// PodIP returns the IP address of the pod. This is read from the environment variable POD_IP.
	PodIP = sync.OnceValue(func() string {
		podIP := os.Getenv(envPodIP)
		if podIP != "" {
			return podIP
		}

		// Fallback to an empty string if the environment variable is not set
		return ""
	})

	// NodeName returns the name of the node on which the pod is running.
	NodeName = sync.OnceValue(func() string {
		nodeName := os.Getenv(envNodeName)
		if nodeName != "" {
			return nodeName
		}

		// Fallback to an empty string if the environment variable is not set
		return ""
	})
)

var (
	// DeployedNamespace returns the namespace in which the pod is deployed. This is read from the Kubernetes namespace file.
	DeployedNamespace = sync.OnceValue(func() string {
		got, err := os.ReadFile(kubernetesNamespacePath)
		if err != nil {
			return "default" // Fallback to default namespace if the file is not found
		}
		return strings.TrimSpace(string(got))
	})

	// IsInCluster checks if the code is running inside a Kubernetes cluster by checking the existence of the service account file.
	IsInCluster = sync.OnceValue(func() bool {
		_, err := os.Stat(kubernetesNamespacePath)
		return err == nil
	})

	// KubernetesService returns the full address (host:port) of the Kubernetes service.
	// These are read from environment variables with fallback to localhost:443.
	KubernetesService = sync.OnceValue(func() string {
		host := os.Getenv(envKubernetesHost)
		if host == "" {
			host = "localhost"
		}

		port := os.Getenv(envKubernetesPort)
		if port == "" {
			port = "443"
		}
		return net.JoinHostPort(host, port)
	})
)

var (
	// ServiceAccountName returns the name of the service account used by the pod. This is read from the environment variable SERVICE_ACCOUNT_NAME.
	ServiceAccountName = sync.OnceValue(func() string {
		serviceAccountName := os.Getenv(envServiceAccountName)
		if serviceAccountName != "" {
			return serviceAccountName
		}

		return "default" // Fallback to default service account if the environment variable is not set
	})

	// ServiceAccountToken returns the token used by the service account. This is read from the Kubernetes service account file.
	ServiceAccountToken = sync.OnceValue(func() string {
		got, err := os.ReadFile(kubernetesServiceAccountTokenPath)
		if err != nil {
			return "" // Token file not found or not readable
		}
		return strings.TrimSpace(string(got))
	})
)
