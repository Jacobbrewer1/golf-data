package vaulty

import hashiVault "github.com/hashicorp/vault/api"

func CipherTextFromSecret(transitEncryptSecret *hashiVault.Secret) string {
	switch {
	case transitEncryptSecret == nil:
		return ""
	case transitEncryptSecret.Data == nil:
		return ""
	case transitEncryptSecret.Data[TransitKeyCipherText] == nil:
		return ""
	}

	ct, ok := transitEncryptSecret.Data[TransitKeyCipherText].(string)
	if !ok {
		return ""
	}

	return ct
}
