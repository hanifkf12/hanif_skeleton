package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestJWTRejectsUnexpectedAlgorithm(t *testing.T) {
	service, err := NewJWT(Config{SecretKey: "test-secret", Issuer: "test"})
	require.NoError(t, err)

	claims := Claims{
		UserID:   1,
		Username: "alice",
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "test",
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS384, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	_, err = service.Parse(signed)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestRefreshRejectsExpiredToken(t *testing.T) {
	service, err := NewJWT(Config{SecretKey: "test-secret", Issuer: "test", Expiry: -time.Second})
	require.NoError(t, err)
	token, err := service.Generate(Claims{UserID: 1, Username: "alice"})
	require.NoError(t, err)

	_, err = service.Refresh(token)
	require.ErrorIs(t, err, ErrTokenExpired)
}

func TestGenerateRequiresIdentityClaims(t *testing.T) {
	service, err := NewJWT(Config{SecretKey: "test-secret"})
	require.NoError(t, err)

	_, err = service.Generate(Claims{})
	require.ErrorIs(t, err, ErrMissingClaims)
}
