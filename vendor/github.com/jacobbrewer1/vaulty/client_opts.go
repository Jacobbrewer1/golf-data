package vaulty

import (
	"context"
	"errors"
	"log/slog"
	"os"

	hashiVault "github.com/hashicorp/vault/api"
	kubernetesAuth "github.com/hashicorp/vault/api/auth/kubernetes"
)

type ClientOption func(c *client) error

// WithContext sets the context for the client.
func WithContext(ctx context.Context) ClientOption {
	return func(c *client) error {
		c.ctx = ctx
		return nil
	}
}

// WithLogger sets the logger for the client.
func WithLogger(l *slog.Logger) ClientOption {
	return func(c *client) error {
		c.l = l
		return nil
	}
}

// WithGeneratedVaultClient creates a vault client with the given address.
//
// Deprecated: Use WithAddr instead for the same effect.
func WithGeneratedVaultClient(vaultAddress string) ClientOption {
	return WithAddr(vaultAddress)
}

// WithAddr sets the address for the client.
func WithAddr(addr string) ClientOption {
	return func(c *client) error {
		c.config.Address = addr
		return nil
	}
}

// WithConfig sets the config for the client.
func WithConfig(config *hashiVault.Config) ClientOption {
	return func(c *client) error {
		c.config = config
		return nil
	}
}

// WithTokenAuth sets the token for the client.
func WithTokenAuth(token string) ClientOption {
	return func(c *client) error {
		if token == "" {
			return errors.New("token is empty")
		}

		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			return tokenLogin(v, token)
		}
		return nil
	}
}

// WithAppRoleAuth sets the AppRole authentication method for the client.
func WithAppRoleAuth(roleID, secretID string) ClientOption {
	return func(c *client) error {
		if roleID == "" {
			return errors.New("roleID is empty")
		} else if secretID == "" {
			return errors.New("secretID is empty")
		}

		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			sec, err := appRoleLogin(v, roleID, secretID)
			if err != nil {
				return nil, err
			}
			go c.renewAuthInfo()
			return sec, nil
		}
		return nil
	}
}

// WithUserPassAuth sets the UserPass authentication method for the client.
func WithUserPassAuth(username, password string) ClientOption {
	return func(c *client) error {
		if username == "" {
			return errors.New("username is empty")
		} else if password == "" {
			return errors.New("password is empty")
		}

		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			sec, err := userPassLogin(v, username, password)
			if err != nil {
				return nil, err
			}
			go c.renewAuthInfo()
			return sec, nil
		}
		return nil
	}
}

// WithKvv2Mount sets the KVv2 mount point for the client.
func WithKvv2Mount(mount string) ClientOption {
	return func(c *client) error {
		c.kvv2Mount = mount
		return nil
	}
}

// WithKubernetesServiceAccountAuth sets the Kubernetes authentication method for the client using a service account.
func WithKubernetesServiceAccountAuth(roleName string) ClientOption {
	return func(c *client) error {
		if roleName == "" {
			return errors.New("role name is empty")
		}

		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			sec, err := kubernetesLogin(v, roleName, kubernetesAuth.WithServiceAccountTokenPath(kubernetesServiceAccountTokenPath))
			if err != nil {
				return nil, err
			}

			go c.renewAuthInfo()

			return sec, nil
		}
		return nil
	}
}

// WithKubernetesAuthFromEnv sets the Kubernetes authentication method for the client using a service account from environment variables.
func WithKubernetesAuthFromEnv() ClientOption {
	return func(c *client) error {
		roleFromEnv := os.Getenv(envServiceAccountName)
		if roleFromEnv == "" {
			return errors.New("role name is not set in environment variable " + envServiceAccountName)
		}

		return WithKubernetesServiceAccountAuth(roleFromEnv)(c)
	}
}

// WithKubernetesAuth sets the Kubernetes authentication method for the client using the role and token provided.
func WithKubernetesAuth(role, token string) ClientOption {
	return func(c *client) error {
		if role == "" {
			return errors.New("role name is empty")
		} else if token == "" {
			return errors.New("token is empty")
		}

		c.auth = func(v *hashiVault.Client) (*hashiVault.Secret, error) {
			sec, err := kubernetesLogin(v, role, kubernetesAuth.WithServiceAccountToken(token))
			if err != nil {
				return nil, err
			}

			go c.renewAuthInfo()

			return sec, nil
		}
		return nil
	}
}
