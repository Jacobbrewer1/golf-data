package repositories

import (
	"context"
	"log/slog"

	hashiVault "github.com/hashicorp/vault/api"
	"github.com/jacobbrewer1/vaulty"
	"github.com/spf13/viper"
)

type ConnectorOption func(c *databaseConnector)

// WithConnectorLogger sets the logger for the client.
func WithConnectorLogger(l *slog.Logger) ConnectorOption {
	return func(c *databaseConnector) {
		c.l = l
	}
}

// WithVaultClient sets the vault client for the client.
func WithVaultClient(client vaulty.Client) ConnectorOption {
	return func(c *databaseConnector) {
		c.client = client
	}
}

// WithViper sets the viper configuration for the client.
func WithViper(v *viper.Viper) ConnectorOption {
	return func(c *databaseConnector) {
		c.vip = v
	}
}

// WithCurrentSecrets sets the current secrets for the client.
func WithCurrentSecrets(secrets *hashiVault.Secret) ConnectorOption {
	return func(c *databaseConnector) {
		c.currentSecrets = secrets
	}
}

// WithContext sets the context for the client.
func WithContext(ctx context.Context) ConnectorOption {
	return func(c *databaseConnector) {
		c.ctx = ctx
	}
}
