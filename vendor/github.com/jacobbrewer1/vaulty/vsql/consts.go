package vsql

const (
	loggingKeyError = "err"

	configKeyVault             = "vault"
	configKeyDatabase          = "database"
	configKeyVaultDatabase     = configKeyVault + "." + configKeyDatabase
	configKeyVaultDatabaseRole = configKeyVaultDatabase + ".role"
	configKeyVaultDatabasePath = configKeyVaultDatabase + ".path"
	configKeyDatabaseHost      = configKeyDatabase + ".host"
	configKeyDatabaseName      = configKeyDatabase + ".schema"

	secretKeyDatabaseUsername = "username"
	secretKeyDatabasePassword = "password"
)
