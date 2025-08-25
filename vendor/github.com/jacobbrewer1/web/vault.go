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
// Note:
//   - This is defined as a variable to allow users of the package to override it if needed.
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
