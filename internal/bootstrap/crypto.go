package bootstrap

import (
	"log"

	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/crypto"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// RegistryCrypto creates and returns a crypto instance based on configuration
func RegistryCrypto(cfg *config.Config) crypto.Crypto {
	lf := logger.NewFields("RegistryCrypto")

	encryptionKey := cfg.Crypto.EncryptionKey
	if encryptionKey == "" {
		log.Fatal("ENCRYPTION_KEY is required. Generate one using: openssl rand -base64 32")
	}

	cryptoInstance, err := crypto.NewCrypto(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize crypto: %v", err)
	}

	logger.Info("Crypto initialized successfully", lf)
	return cryptoInstance
}

// RegistryBcryptHasher creates and returns a bcrypt hasher instance
func RegistryBcryptHasher(cfg *config.Config) *crypto.BcryptHasher {
	lf := logger.NewFields("RegistryBcryptHasher")

	cost := cfg.Crypto.BcryptCost
	if cost == 0 {
		cost = 10 // Default cost
	}

	lf.Append(logger.Any("cost", cost))

	hasher := crypto.NewBcryptHasher(cost)

	logger.Info("Bcrypt hasher initialized successfully", lf)
	return hasher
}
