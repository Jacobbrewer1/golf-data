package vaulty

const (
	loggingKeyError         = "err"
	loggingKeySecretName    = "secret"
	loggingKeyResult        = "result"
	loggingKeyRenewedAt     = "renewed_at"
	loggingKeyLeaseDuration = "lease_duration"

	pathKeyTransitDecrypt = "decrypt"
	pathKeyTransitEncrypt = "encrypt"

	TransitKeyCipherText = "ciphertext"
	TransitKeyPlainText  = "plaintext"

	envKubernetesRole  = "KUBERNETES_ROLE"  // nolint:gosec // This is detected as a secret
	envKubernetesToken = "KUBERNETES_TOKEN" // nolint:gosec // This is detected as a secret

	// KubernetesServiceAccountTokenPath is the path to the Kubernetes service account token.
	kubernetesServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token" // nolint:gosec // This is detected as a secret
)
