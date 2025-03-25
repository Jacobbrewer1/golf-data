package web

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jacobbrewer1/vaulty"
	"github.com/spf13/viper"
)

const (
	// defaultVaultAddr is the default address of the Vault server
	defaultVaultAddr = "http://vault-active.vault:8200"
)

var VaultClient = func(ctx context.Context, l *slog.Logger, v *viper.Viper) (vaulty.Client, error) {
	addr := v.GetString("vault.address")
	if addr == "" {
		addr = defaultVaultAddr
	}

	vc, err := vaulty.NewClient(
		vaulty.WithContext(ctx),
		vaulty.WithAddr(addr),
		vaulty.WithKubernetesAuthDefault(),
		vaulty.WithKvv2Mount(v.GetString("vault.kvv2_mount")),
		vaulty.WithLogger(l),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating vault client: %w", err)
	}

	return vc, nil
}
