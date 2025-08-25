package k8s

import (
	"net"
	"os"
	"strings"
	"sync"
)

const (
	// kubernetesServiceAccountPath specifies the file path to the Kubernetes service account directory.
	kubernetesServiceAccountPath = "/var/run/secrets/kubernetes.io/serviceaccount"

	// kubernetesNamespacePath specifies the file path to the Kubernetes namespace file.
	kubernetesNamespacePath = kubernetesServiceAccountPath + "/namespace"

	// kubernetesServiceAccountTokenPath specifies the file path to the Kubernetes service account token file.
	kubernetesServiceAccountTokenPath = kubernetesServiceAccountPath + "/token"
)

const (
	// envHostname is the environment variable that holds the pod's hostname.
	envHostname = "HOSTNAME"

	// envPodIP is the environment variable that holds the pod's IP address.
	envPodIP = "POD_IP"

	// envServiceAccountName is the environment variable that holds the name of the service account used by the pod.
	envServiceAccountName = "SERVICE_ACCOUNT_NAME"

	// envKubernetesHost is the environment variable that holds the Kubernetes service host address.
	envKubernetesHost = "KUBERNETES_SERVICE_HOST"

	// envKubernetesPort is the environment variable that holds the Kubernetes service port.
	envKubernetesPort = "KUBERNETES_SERVICE_PORT"

	// envNodeName is the environment variable that holds the name of the node on which the pod is running.
	envNodeName = "NODE_NAME"
)

var (
	// PodName returns the name of the pod.
	//
	// Kubernetes pods are typically identified by their hostname, which is set to the pod name.
	// This is accessible via the `os.Hostname()` function or the `HOSTNAME` environment variable.
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

	// PodIP returns the IP address of the pod.
	//
	// This function retrieves the pod's IP address by reading the value of the
	// environment variable `POD_IP`. If the environment variable is not set,
	// it falls back to returning an empty string.
	//
	// Note: The `POD_IP` environment variable is not a standard Kubernetes
	// environment variable. This can be set using the downward API in the pod spec.
	PodIP = sync.OnceValue(func() string {
		podIP := os.Getenv(envPodIP)
		if podIP != "" {
			return podIP
		}

		// Fallback to an empty string if the environment variable is not set
		return ""
	})

	// NodeName returns the name of the node on which the pod is running.
	//
	// This function retrieves the node name by reading the value of the
	// environment variable `NODE_NAME`. If the environment variable is not set,
	// it falls back to returning an empty string.
	//
	// Note: The `NODE_NAME` environment variable is not a standard Kubernetes
	// environment variable. This can be set using the downward API in the pod spec.
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
	// DeployedNamespace returns the namespace in which the pod is deployed.
	//
	// This function reads the namespace from the Kubernetes namespace file stored in
	// the service account directory. If the file cannot be read (e.g., it does not exist
	// or is not accessible), the function falls back to returning the default namespace ("default").
	DeployedNamespace = sync.OnceValue(func() string {
		got, err := os.ReadFile(kubernetesNamespacePath)
		if err != nil {
			return "default" // Fallback to default namespace if the file is not found
		}
		return strings.TrimSpace(string(got))
	})

	// IsInCluster checks if the code is running inside a Kubernetes cluster.
	//
	// This function determines whether the application is running inside a Kubernetes
	// cluster by checking for the existence of the Kubernetes namespace file.
	IsInCluster = sync.OnceValue(func() bool {
		_, err := os.Stat(kubernetesNamespacePath)
		return err == nil
	})

	// KubernetesService returns the full address (host:port) of the Kubernetes service.
	//
	// This function constructs the Kubernetes service address by reading the host and port
	// from the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT`.
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
	// ServiceAccountName returns the name of the service account used by the pod.
	//
	// This function retrieves the service account name by reading the value of the
	// environment variable `SERVICE_ACCOUNT_NAME`.
	//
	// Note: The `SERVICE_ACCOUNT_NAME` environment variable is not a standard Kubernetes
	// environment variable. This can be set using the downward API in the pod spec.
	ServiceAccountName = sync.OnceValue(func() string {
		serviceAccountName := os.Getenv(envServiceAccountName)
		if serviceAccountName != "" {
			return serviceAccountName
		}

		return "default" // Fallback to default service account if the environment variable is not set
	})

	// ServiceAccountToken returns the token used by the service account.
	//
	// This function reads the service account token from the Kubernetes service account token file.
	// If the file cannot be read (e.g., it does not exist or is not accessible), the function falls
	// back to returning an empty string.
	ServiceAccountToken = sync.OnceValue(func() string {
		got, err := os.ReadFile(kubernetesServiceAccountTokenPath)
		if err != nil {
			return "" // Token file not found or not readable
		}
		return strings.TrimSpace(string(got))
	})
)
