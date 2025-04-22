package vaulty

const (
	maxInt = int(^uint(0) >> 1) // maximum value for int

	loggingKeyError         = "err"
	loggingKeySecretName    = "secret"
	loggingKeyResult        = "result"
	loggingKeyRenewedAt     = "renewed_at"
	loggingKeyLeaseDuration = "lease_duration"

	pathKeyTransitDecrypt = "decrypt"
	pathKeyTransitEncrypt = "encrypt"

	TransitKeyCipherText = "ciphertext"
	TransitKeyPlainText  = "plaintext"

	envServiceAccountName = "SERVICE_ACCOUNT_NAME" // nolint:gosec // This is detected as a secret

	// KubernetesServiceAccountTokenPath is the path to the Kubernetes service account token.
	kubernetesServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token" // nolint:gosec // This is detected as a secret
)
