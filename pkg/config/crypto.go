package config

// Crypto holds cryptography configuration
type Crypto struct {
	EncryptionKey string `mapstructure:"ENCRYPTION_KEY"` // Secret key for AES encryption
	BcryptCost    int    `mapstructure:"BCRYPT_COST"`    // Bcrypt cost factor (4-31, default 10)
}
