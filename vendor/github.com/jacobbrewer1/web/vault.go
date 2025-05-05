package web

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/viper"

	"github.com/jacobbrewer1/vaulty"
	"github.com/jacobbrewer1/web/k8s"
)

// defaultVaultAddr is the default address of the Vault server.
//
// This constant defines the default URL for the active Vault server,
// which is used when no custom address is provided in the configuration.
const (
	// defaultVaultAddr is the default address of the Vault server
	defaultVaultAddr = "http://vault-active.vault:8200"
)

// VaultClient creates and configures a new Vault client.
//
// This function initializes a Vault client using the provided context, logger, and viper configuration.
// It retrieves the Vault server address from the configuration, falling back to the default address
// if none is specified. The client is configured with Kubernetes service account authentication,
// a KVv2 mount point, and logging.
//
// Parameters:
//   - ctx: The context.Context used for the Vault client.
//   - l: A pointer to an slog.Logger instance for logging purposes.
//   - v: A pointer to a viper.Viper instance for configuration management.
//
// Returns:
//   - vaulty.Client: The configured Vault client.
//   - error: An error if the Vault client setup fails.
//
// Behavior:
//   - Retrieves the Vault server address from the viper configuration.
//   - Uses the defaultVaultAddr if no address is provided.
//   - Configures the Vault client with the provided context, address, Kubernetes service account authentication,
//     KVv2 mount point, and logger.
//   - Returns the configured client or an error if the setup fails.
var VaultClient = func(ctx context.Context, l *slog.Logger, v *viper.Viper) (vaulty.Client, error) {
	addr := v.GetString("vault.address")
	if addr == "" {
		addr = defaultVaultAddr
	}

	vc, err := vaulty.NewClient(
		vaulty.WithContext(ctx),
		vaulty.WithAddr(addr),
		vaulty.WithKubernetesServiceAccountAuth(k8s.ServiceAccountName()),
		vaulty.WithKvv2Mount(v.GetString("vault.kvv2_mount")),
		vaulty.WithLogger(l),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating vault client: %w", err)
	}

	return vc, nil
}
