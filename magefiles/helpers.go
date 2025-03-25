//go:build mage

package main

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

func getLocalVaultClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = "http://localhost:8200"

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating vault client: %w", err)
	}

	client.SetToken("root")

	return client, nil
}
