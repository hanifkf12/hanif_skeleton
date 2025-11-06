package bootstrap

import (
	"log"
	"time"

	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/jwt"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// RegistryJWT creates and returns a JWT instance based on configuration
func RegistryJWT(cfg *config.Config) jwt.JWT {
	lf := logger.NewFields("RegistryJWT")

	secretKey := cfg.JWT.SecretKey
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY is required. Generate one using: openssl rand -base64 32")
	}

	issuer := cfg.JWT.Issuer
	if issuer == "" {
		issuer = "hanif-skeleton"
	}

	expiry := cfg.JWT.Expiry
	if expiry == 0 {
		expiry = 24 * time.Hour // Default 24 hours
	}

	lf.Append(logger.Any("issuer", issuer))
	lf.Append(logger.Any("expiry", expiry.String()))

	jwtInstance, err := jwt.NewJWT(jwt.Config{
		SecretKey: secretKey,
		Issuer:    issuer,
		Expiry:    expiry,
	})

	if err != nil {
		log.Fatalf("Failed to initialize JWT: %v", err)
	}

	logger.Info("JWT initialized successfully", lf)
	return jwtInstance
}
