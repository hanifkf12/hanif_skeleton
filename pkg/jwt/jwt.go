package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidSignMethod = errors.New("invalid signing method")
	ErrMissingClaims     = errors.New("missing claims")
)

// JWT handles JWT token operations
type JWT interface {
	// Generate generates a new JWT token with claims
	Generate(claims Claims) (string, error)

	// Parse parses and validates a JWT token
	Parse(tokenString string) (*Claims, error)

	// Refresh refreshes an existing token with new expiry
	Refresh(tokenString string) (string, error)

	// Validate validates a token without parsing claims
	Validate(tokenString string) error
}

// Claims represents JWT claims
type Claims struct {
	UserID   int64             `json:"user_id"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Role     string            `json:"role"`
	Extra    map[string]string `json:"extra,omitempty"`
	jwtlib.RegisteredClaims
}

// jwtImpl implements JWT interface
type jwtImpl struct {
	secretKey []byte
	issuer    string
	expiry    time.Duration
}

// Config holds JWT configuration
type Config struct {
	SecretKey string        // Secret key for signing
	Issuer    string        // Token issuer
	Expiry    time.Duration // Token expiry duration
}

// NewJWT creates a new JWT instance
func NewJWT(config Config) (JWT, error) {
	if config.SecretKey == "" {
		return nil, errors.New("secret key is required")
	}

	if config.Expiry == 0 {
		config.Expiry = 24 * time.Hour // Default 24 hours
	}

	if config.Issuer == "" {
		config.Issuer = "hanif-skeleton"
	}

	return &jwtImpl{
		secretKey: []byte(config.SecretKey),
		issuer:    config.Issuer,
		expiry:    config.Expiry,
	}, nil
}

// Generate generates a new JWT token
func (j *jwtImpl) Generate(claims Claims) (string, error) {
	now := time.Now()

	// Set registered claims
	claims.Issuer = j.issuer
	claims.IssuedAt = jwtlib.NewNumericDate(now)
	claims.ExpiresAt = jwtlib.NewNumericDate(now.Add(j.expiry))
	claims.NotBefore = jwtlib.NewNumericDate(now)

	// Create token with claims
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Parse parses and validates a JWT token
func (j *jwtImpl) Parse(tokenString string) (*Claims, error) {
	// Parse token
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignMethod
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// Refresh refreshes an existing token with new expiry
func (j *jwtImpl) Refresh(tokenString string) (string, error) {
	// Parse existing token
	claims, err := j.Parse(tokenString)
	if err != nil {
		// If token is expired, we can still refresh it
		if !errors.Is(err, ErrTokenExpired) {
			return "", err
		}
		// Parse without validation for refresh
		token, _ := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
			return j.secretKey, nil
		})
		claims, _ = token.Claims.(*Claims)
	}

	// Generate new token with same claims
	newToken, err := j.Generate(*claims)
	if err != nil {
		return "", err
	}

	return newToken, nil
}

// Validate validates a token without parsing claims
func (j *jwtImpl) Validate(tokenString string) error {
	_, err := j.Parse(tokenString)
	return err
}

// GetUserID extracts user ID from claims
func (c *Claims) GetUserID() int64 {
	return c.UserID
}

// GetUsername extracts username from claims
func (c *Claims) GetUsername() string {
	return c.Username
}

// GetEmail extracts email from claims
func (c *Claims) GetEmail() string {
	return c.Email
}

// GetRole extracts role from claims
func (c *Claims) GetRole() string {
	return c.Role
}

// IsAdmin checks if user is admin
func (c *Claims) IsAdmin() bool {
	return c.Role == "admin" || c.Role == "superadmin"
}

// HasRole checks if user has specific role
func (c *Claims) HasRole(role string) bool {
	return c.Role == role
}
