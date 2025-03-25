package vaulty

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	hashiVault "github.com/hashicorp/vault/api"
	kubernetesAuth "github.com/hashicorp/vault/api/auth/kubernetes"
)

type ClientOption func(c *client)

// WithContext sets the context for the client.
func WithContext(ctx context.Context) ClientOption {
	return func(c *client) {
		c.ctx = ctx
	}
}

// WithLogger sets the logger for the client.
func WithLogger(l *slog.Logger) ClientOption {
	return func(c *client) {
		c.l = l
	}
}

// WithGeneratedVaultClient creates a vault client with the given address.
//
// Deprecated: Use WithAddr instead for the same effect.
func WithGeneratedVaultClient(vaultAddress string) ClientOption {
	return WithAddr(vaultAddress)
}

func WithAddr(addr string) ClientOption {
	return func(c *client) {
		c.config.Address = addr
	}
}

func WithConfig(config *hashiVault.Config) ClientOption {
	return func(c *client) {
		c.config = config
	}
}

func WithTokenAuth(token string) ClientOption {
	return func(c *client) {
		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			return tokenLogin(v, token)
		}
	}
}

func WithAppRoleAuth(roleID, secretID string) ClientOption {
	return func(c *client) {
		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			sec, err := appRoleLogin(v, roleID, secretID)
			if err != nil {
				return nil, err
			}
			go c.renewAuthInfo()
			return sec, nil
		}
	}
}

func WithUserPassAuth(username, password string) ClientOption {
	return func(c *client) {
		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			sec, err := userPassLogin(v, username, password)
			if err != nil {
				return nil, err
			}
			go c.renewAuthInfo()
			return sec, nil
		}
	}
}

func WithKvv2Mount(mount string) ClientOption {
	return func(c *client) {
		c.kvv2Mount = mount
	}
}

func WithKubernetesAuthDefault() ClientOption {
	return func(c *client) {
		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			role := os.Getenv(envKubernetesRole)
			if role == "" {
				return nil, fmt.Errorf("%s environment variable not set", envKubernetesRole)
			}

			sec, err := kubernetesLogin(v, role, kubernetesAuth.WithServiceAccountTokenPath(kubernetesServiceAccountTokenPath))
			if err != nil {
				return nil, err
			}

			go c.renewAuthInfo()

			return sec, nil
		}
	}
}

func WithKubernetesAuthFromEnv() ClientOption {
	return func(c *client) {
		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			role := os.Getenv(envKubernetesRole)
			if role == "" {
				return nil, fmt.Errorf("%s environment variable not set", envKubernetesRole)
			}

			sec, err := kubernetesLogin(v, role, kubernetesAuth.WithServiceAccountTokenEnv(envKubernetesToken))
			if err != nil {
				return nil, err
			}

			go c.renewAuthInfo()

			return sec, nil
		}
	}
}

func WithKubernetesAuth(role, token string) ClientOption {
	return func(c *client) {
		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			sec, err := kubernetesLogin(v, role, kubernetesAuth.WithServiceAccountToken(token))
			if err != nil {
				return nil, err
			}

			go c.renewAuthInfo()

			return sec, nil
		}
	}
}
