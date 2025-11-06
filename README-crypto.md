# Crypto Package Documentation

## Overview

Crypto package menyediakan utility untuk **encryption, decryption, dan hashing** data sensitif. Package ini menggunakan **AES-256-GCM** untuk encryption dan **bcrypt** untuk password hashing, mengikuti best practices security modern.

## Architecture

```
┌─────────────────────────────────────┐
│         UseCase Layer               │
├─────────────────────────────────────┤
│    Crypto Interface (Contract)      │  ← Abstraction
├─────────────────────────────────────┤
│    Implementation                   │
│    ├─ AES-256-GCM Encryption       │
│    ├─ SHA-256 Hashing              │
│    └─ Bcrypt Password Hashing      │
└─────────────────────────────────────┘
```

## Features

### ✅ AES-256-GCM Encryption
- **Symmetric encryption** with authentication
- **PBKDF2 key derivation** for added security
- **Random nonce** for each encryption
- **Base64 encoding** for easy storage/transmission

### ✅ SHA-256 Hashing
- **One-way hashing** for data fingerprinting
- **Deterministic** - same input always produces same hash
- **Base64 encoded** output

### ✅ Bcrypt Password Hashing
- **Adaptive hashing** for passwords
- **Salted automatically**
- **Configurable cost** factor
- **Secure comparison** function

## Crypto Interface

```go
type Crypto interface {
    // Encrypt encrypts plaintext and returns base64 encoded ciphertext
    Encrypt(plaintext string) (string, error)

    // Decrypt decrypts base64 encoded ciphertext and returns plaintext
    Decrypt(ciphertext string) (string, error)

    // EncryptBytes encrypts byte data and returns encrypted bytes
    EncryptBytes(data []byte) ([]byte, error)

    // DecryptBytes decrypts encrypted bytes and returns plaintext bytes
    DecryptBytes(data []byte) ([]byte, error)

    // Hash creates a SHA256 hash of the input
    Hash(data string) string

    // CompareHash compares data with a hash
    CompareHash(data string, hash string) bool
}
```

## Configuration

### Environment Variables

Add to `.env`:

```bash
# Crypto Configuration
# Generate key: openssl rand -base64 32
ENCRYPTION_KEY=your-secret-encryption-key-here
BCRYPT_COST=10
```

### Generate Secure Key

```bash
# Generate a 32-byte random key (recommended)
openssl rand -base64 32

# Example output:
# k8/JzQ7+FjKxN1mL9pW3vR5tY2nU6hG0iS4eA8bC7dE=
```

**⚠️ IMPORTANT:**
- Use a different key for each environment (dev/staging/prod)
- Never commit encryption keys to version control
- Store keys securely (use secrets manager in production)
- Rotate keys periodically

### Config Struct

File: `pkg/config/crypto.go`

```go
type Crypto struct {
    EncryptionKey string // AES encryption key
    BcryptCost    int    // 4-31, default 10
}
```

## Bootstrap Registry

File: `internal/bootstrap/crypto.go`

```go
// Initialize crypto
crypto := bootstrap.RegistryCrypto(cfg)

// Initialize bcrypt hasher
hasher := bootstrap.RegistryBcryptHasher(cfg)
```

## Usage Examples

### 1. Basic Encryption/Decryption

```go
package main

import (
    "github.com/hanifkf12/hanif_skeleton/pkg/crypto"
)

func example() {
    // Initialize
    cryptoInstance, _ := crypto.NewCrypto("your-secret-key")

    // Encrypt
    plaintext := "sensitive data"
    encrypted, err := cryptoInstance.Encrypt(plaintext)
    if err != nil {
        panic(err)
    }
    fmt.Println("Encrypted:", encrypted)
    // Output: Encrypted: f8D7Kg9Lm2Np3Qr5Ts6Vw8Yx0Az1Bc2De3...

    // Decrypt
    decrypted, err := cryptoInstance.Decrypt(encrypted)
    if err != nil {
        panic(err)
    }
    fmt.Println("Decrypted:", decrypted)
    // Output: Decrypted: sensitive data
}
```

### 2. Encrypt/Decrypt Bytes

```go
func encryptBytes() {
    cryptoInstance, _ := crypto.NewCrypto("your-secret-key")

    // Encrypt byte data
    data := []byte("binary data here")
    encrypted, err := cryptoInstance.EncryptBytes(data)
    
    // Decrypt
    decrypted, err := cryptoInstance.DecryptBytes(encrypted)
}
```

### 3. Hashing for Data Integrity

```go
func hashExample() {
    cryptoInstance, _ := crypto.NewCrypto("your-secret-key")

    // Create hash
    data := "important data"
    hash := cryptoInstance.Hash(data)
    fmt.Println("Hash:", hash)

    // Verify data integrity
    if cryptoInstance.CompareHash(data, hash) {
        fmt.Println("Data is valid!")
    }
}
```

### 4. Password Hashing with Bcrypt

```go
func passwordExample() {
    hasher := crypto.NewBcryptHasher(10)

    // Hash password
    password := "user_password123"
    hashedPassword, err := hasher.HashPassword(password)
    if err != nil {
        panic(err)
    }
    
    // Store hashedPassword in database
    fmt.Println("Hashed:", hashedPassword)

    // Verify password
    isValid := hasher.ComparePassword(password, hashedPassword)
    if isValid {
        fmt.Println("Password is correct!")
    }
}
```

### 5. Generate Random Keys

```go
func generateKeys() {
    // Generate random key
    randomKey, _ := crypto.GenerateRandomKey(32)
    fmt.Println("Random Key:", randomKey)

    // Generate secret key for AES-256
    secretKey, _ := crypto.GenerateSecretKey()
    fmt.Println("Secret Key:", secretKey)
}
```

### 6. Encrypt Sensitive User Data

```go
type User struct {
    ID              int
    Username        string
    Email           string
    EncryptedPhone  string  // Encrypted
    EncryptedSSN    string  // Encrypted
}

func encryptUserData(cryptoInstance crypto.Crypto, user *User, phone, ssn string) error {
    // Encrypt phone
    encryptedPhone, err := cryptoInstance.Encrypt(phone)
    if err != nil {
        return err
    }
    user.EncryptedPhone = encryptedPhone

    // Encrypt SSN
    encryptedSSN, err := cryptoInstance.Encrypt(ssn)
    if err != nil {
        return err
    }
    user.EncryptedSSN = encryptedSSN

    return nil
}

func decryptUserData(cryptoInstance crypto.Crypto, user *User) (phone, ssn string, err error) {
    // Decrypt phone
    phone, err = cryptoInstance.Decrypt(user.EncryptedPhone)
    if err != nil {
        return "", "", err
    }

    // Decrypt SSN
    ssn, err = cryptoInstance.Decrypt(user.EncryptedSSN)
    if err != nil {
        return "", "", err
    }

    return phone, ssn, nil
}
```

### 7. HTTP Endpoints (Example UseCase)

See: `internal/usecase/encrypt_decrypt.go`

**Encrypt Endpoint:**
```bash
curl -X POST http://localhost:9000/encrypt \
  -H "Content-Type: application/json" \
  -d '{"data": "my secret data"}'

# Response:
{
  "code": 200,
  "data": {
    "encrypted_data": "f8D7Kg9Lm2Np3Qr5Ts6Vw8Yx0...",
    "hash": "9Kj8Hg7Fd6Sa5Df4Gh3Jk2..."
  }
}
```

**Decrypt Endpoint:**
```bash
curl -X POST http://localhost:9000/decrypt \
  -H "Content-Type: application/json" \
  -d '{"encrypted_data": "f8D7Kg9Lm2Np3Qr5Ts6Vw8Yx0..."}'

# Response:
{
  "code": 200,
  "data": {
    "data": "my secret data"
  }
}
```

## Integration with Router

File: `internal/router/router.go`

```go
func (rtr *router) Route() {
    db := bootstrap.RegistryDatabase(rtr.cfg, false)
    cryptoInstance := bootstrap.RegistryCrypto(rtr.cfg)
    hasher := bootstrap.RegistryBcryptHasher(rtr.cfg)

    // Encrypt/Decrypt endpoints
    encryptUseCase := usecase.NewEncryptData(cryptoInstance)
    rtr.fiber.Post("/encrypt", rtr.handle(
        handler.HttpRequest,
        encryptUseCase,
    ))

    decryptUseCase := usecase.NewDecryptData(cryptoInstance)
    rtr.fiber.Post("/decrypt", rtr.handle(
        handler.HttpRequest,
        decryptUseCase,
    ))
}
```

## Use Cases

### 1. **Encrypt PII (Personally Identifiable Information)**

```go
// Before storing in database
func createUser(user *User, cryptoInstance crypto.Crypto) error {
    // Encrypt sensitive fields
    encryptedEmail, _ := cryptoInstance.Encrypt(user.Email)
    encryptedPhone, _ := cryptoInstance.Encrypt(user.Phone)
    encryptedSSN, _ := cryptoInstance.Encrypt(user.SSN)

    // Store encrypted values
    user.Email = encryptedEmail
    user.Phone = encryptedPhone
    user.SSN = encryptedSSN

    return db.Save(user)
}

// When retrieving
func getUser(id int, cryptoInstance crypto.Crypto) (*User, error) {
    user, _ := db.Find(id)

    // Decrypt sensitive fields
    user.Email, _ = cryptoInstance.Decrypt(user.Email)
    user.Phone, _ = cryptoInstance.Decrypt(user.Phone)
    user.SSN, _ = cryptoInstance.Decrypt(user.SSN)

    return user, nil
}
```

### 2. **Password Storage**

```go
// Registration
func registerUser(username, password string, hasher *crypto.BcryptHasher) error {
    // Hash password
    hashedPassword, err := hasher.HashPassword(password)
    if err != nil {
        return err
    }

    user := User{
        Username: username,
        Password: hashedPassword, // Store hashed, never plaintext
    }

    return db.Save(&user)
}

// Login
func loginUser(username, password string, hasher *crypto.BcryptHasher) (*User, error) {
    user, _ := db.FindByUsername(username)

    // Compare password
    if !hasher.ComparePassword(password, user.Password) {
        return nil, errors.New("invalid credentials")
    }

    return user, nil
}
```

### 3. **API Key Encryption**

```go
func storeAPIKey(userID int, apiKey string, cryptoInstance crypto.Crypto) error {
    // Encrypt API key before storage
    encryptedKey, err := cryptoInstance.Encrypt(apiKey)
    if err != nil {
        return err
    }

    return db.SaveAPIKey(userID, encryptedKey)
}

func getAPIKey(userID int, cryptoInstance crypto.Crypto) (string, error) {
    encryptedKey, _ := db.GetAPIKey(userID)

    // Decrypt when needed
    apiKey, err := cryptoInstance.Decrypt(encryptedKey)
    return apiKey, err
}
```

### 4. **Credit Card Tokenization**

```go
func tokenizeCreditCard(cardNumber string, cryptoInstance crypto.Crypto) (string, error) {
    // Encrypt full card number
    encrypted, err := cryptoInstance.Encrypt(cardNumber)
    if err != nil {
        return "", err
    }

    // Create token (hash for reference)
    token := cryptoInstance.Hash(cardNumber)[:16]

    // Store both
    db.SaveCard(token, encrypted)

    return token, nil
}

func retrieveCreditCard(token string, cryptoInstance crypto.Crypto) (string, error) {
    encrypted, _ := db.GetCardByToken(token)
    return cryptoInstance.Decrypt(encrypted)
}
```

### 5. **Session Data Encryption**

```go
func createEncryptedSession(userID int, data map[string]interface{}, cryptoInstance crypto.Crypto) (string, error) {
    // Serialize session data
    jsonData, _ := json.Marshal(data)

    // Encrypt
    encrypted, err := cryptoInstance.EncryptBytes(jsonData)
    if err != nil {
        return "", err
    }

    // Encode to base64 for cookie storage
    sessionToken := base64.StdEncoding.EncodeToString(encrypted)
    return sessionToken, nil
}

func decryptSession(sessionToken string, cryptoInstance crypto.Crypto) (map[string]interface{}, error) {
    // Decode from base64
    encrypted, _ := base64.StdEncoding.DecodeString(sessionToken)

    // Decrypt
    decrypted, err := cryptoInstance.DecryptBytes(encrypted)
    if err != nil {
        return nil, err
    }

    // Deserialize
    var data map[string]interface{}
    json.Unmarshal(decrypted, &data)

    return data, nil
}
```

## Security Best Practices

### 1. **Key Management**

```bash
# Development
ENCRYPTION_KEY=dev-key-only-for-testing

# Production (use secrets manager)
# AWS Secrets Manager
aws secretsmanager get-secret-value --secret-id prod/encryption-key

# Google Secret Manager
gcloud secrets versions access latest --secret="encryption-key"

# HashiCorp Vault
vault kv get secret/encryption-key
```

### 2. **Never Log Sensitive Data**

```go
// ❌ BAD
logger.Info("User data", logger.Any("ssn", user.SSN))

// ✅ GOOD
logger.Info("User data", logger.Any("user_id", user.ID))
```

### 3. **Encrypt Before Database**

```go
// ✅ GOOD - Encrypt at application layer
encryptedData, _ := crypto.Encrypt(sensitiveData)
db.Save(encryptedData)

// Also consider database-level encryption for defense in depth
```

### 4. **Use Different Keys per Environment**

```bash
# .env.development
ENCRYPTION_KEY=dev-key-abc123

# .env.staging
ENCRYPTION_KEY=staging-key-def456

# .env.production
ENCRYPTION_KEY=prod-key-xyz789
```

### 5. **Rotate Keys Periodically**

```go
// Example key rotation strategy
func rotateEncryptionKey(oldCrypto, newCrypto crypto.Crypto) error {
    users, _ := db.GetAllUsers()

    for _, user := range users {
        // Decrypt with old key
        decrypted, _ := oldCrypto.Decrypt(user.EncryptedData)

        // Re-encrypt with new key
        reencrypted, _ := newCrypto.Encrypt(decrypted)

        // Update database
        user.EncryptedData = reencrypted
        db.Save(&user)
    }

    return nil
}
```

### 6. **Bcrypt Cost Factor**

```go
// Development: Lower cost for speed
BCRYPT_COST=4

// Production: Higher cost for security
BCRYPT_COST=12

// Adjust based on:
// - Available CPU
// - Acceptable latency
// - Security requirements
```

## Performance Considerations

### Encryption Performance

```go
// For small strings (< 1KB): ~0.1ms
crypto.Encrypt("small data")

// For large strings (> 1MB): Consider chunking
func encryptLargeData(data []byte, crypto crypto.Crypto) ([]byte, error) {
    chunkSize := 1024 * 1024 // 1MB chunks
    // Process in chunks...
}
```

### Bcrypt Performance

```go
// Cost 10: ~100ms
// Cost 12: ~400ms
// Cost 14: ~1600ms

// Balance security vs user experience
// For APIs: Cost 10-12
// For critical data: Cost 12-14
```

### Caching Decrypted Data

```go
// Cache decrypted data to avoid repeated decryption
type CachedUser struct {
    ID              int
    DecryptedEmail  string
    DecryptedPhone  string
    CachedAt        time.Time
}

// Invalidate cache after 5 minutes
const CacheTTL = 5 * time.Minute
```

## Error Handling

```go
func handleCryptoError(err error) appctx.Response {
    if err == nil {
        return *appctx.NewResponse().WithCode(200)
    }

    // Check for specific errors
    if err == crypto.ErrInvalidKey {
        return *appctx.NewResponse().
            WithCode(500).
            WithErrors("Invalid encryption configuration")
    }

    if err == crypto.ErrInvalidCiphertext {
        return *appctx.NewResponse().
            WithCode(400).
            WithErrors("Invalid encrypted data")
    }

    if err == crypto.ErrDecryptionFailed {
        return *appctx.NewResponse().
            WithCode(400).
            WithErrors("Failed to decrypt data")
    }

    // Generic error
    return *appctx.NewResponse().
        WithCode(500).
        WithErrors("Encryption error")
}
```

## Testing

### Unit Test Encryption

```go
func TestEncryptDecrypt(t *testing.T) {
    crypto, _ := crypto.NewCrypto("test-key")

    plaintext := "test data"
    
    // Encrypt
    encrypted, err := crypto.Encrypt(plaintext)
    assert.NoError(t, err)
    assert.NotEqual(t, plaintext, encrypted)

    // Decrypt
    decrypted, err := crypto.Decrypt(encrypted)
    assert.NoError(t, err)
    assert.Equal(t, plaintext, decrypted)
}
```

### Test Password Hashing

```go
func TestPasswordHash(t *testing.T) {
    hasher := crypto.NewBcryptHasher(4) // Low cost for tests

    password := "password123"

    // Hash
    hash, err := hasher.HashPassword(password)
    assert.NoError(t, err)

    // Verify correct password
    assert.True(t, hasher.ComparePassword(password, hash))

    // Verify incorrect password
    assert.False(t, hasher.ComparePassword("wrong", hash))
}
```

## Compliance

### GDPR / Data Protection

```go
// Right to be forgotten: Easy data deletion
func deleteUserData(userID int, crypto crypto.Crypto) error {
    // Delete encrypted data from database
    // No need to decrypt before deletion
    return db.DeleteUser(userID)
}

// Right to data portability: Export decrypted data
func exportUserData(userID int, crypto crypto.Crypto) (map[string]string, error) {
    user, _ := db.GetUser(userID)

    return map[string]string{
        "email": crypto.Decrypt(user.EncryptedEmail),
        "phone": crypto.Decrypt(user.EncryptedPhone),
        // Don't export sensitive fields like SSN
    }, nil
}
```

### PCI-DSS (Payment Card Industry)

```go
// Never store full card number unencrypted
func storeCardSecurely(cardNumber string, crypto crypto.Crypto) error {
    // Encrypt full number
    encrypted, _ := crypto.Encrypt(cardNumber)

    // Store last 4 digits separately for display
    last4 := cardNumber[len(cardNumber)-4:]

    db.SaveCard(encrypted, last4)
    return nil
}
```

## Troubleshooting

### Issue: Decryption fails after key change
**Solution**: Use key versioning

```go
type EncryptedData struct {
    Data       string
    KeyVersion int
}

// Decrypt with appropriate key version
func decryptWithVersion(data EncryptedData) (string, error) {
    key := getKeyForVersion(data.KeyVersion)
    crypto, _ := crypto.NewCrypto(key)
    return crypto.Decrypt(data.Data)
}
```

### Issue: Performance slow with bcrypt
**Solution**: Reduce cost or use async processing

```go
// Hash password asynchronously
func registerAsync(username, password string) {
    go func() {
        hasher := crypto.NewBcryptHasher(12)
        hash, _ := hasher.HashPassword(password)
        db.SaveUser(username, hash)
    }()
}
```

### Issue: Encryption key not found
**Solution**: Validate at startup

```go
func validateConfig(cfg *config.Config) error {
    if cfg.Crypto.EncryptionKey == "" {
        return errors.New("ENCRYPTION_KEY is required")
    }
    if len(cfg.Crypto.EncryptionKey) < 32 {
        return errors.New("ENCRYPTION_KEY must be at least 32 characters")
    }
    return nil
}
```

## Summary

Crypto package provides:
- ✅ **AES-256-GCM encryption** for sensitive data
- ✅ **SHA-256 hashing** for data integrity
- ✅ **Bcrypt password hashing** for authentication
- ✅ **Bootstrap integration** for easy setup
- ✅ **Clean Architecture** pattern
- ✅ **Production ready** with proper error handling
- ✅ **Well documented** with real-world examples

**Choose encryption based on use case:**
- **AES-256-GCM**: Two-way encryption (PII, API keys, credit cards)
- **SHA-256**: One-way hashing (data integrity, fingerprinting)
- **Bcrypt**: Password hashing (user authentication)

---

**Files:**
- Interface: `pkg/crypto/crypto.go`
- Password Hashing: `pkg/crypto/hasher.go`
- Config: `pkg/config/crypto.go`
- Bootstrap: `internal/bootstrap/crypto.go`
- Example UseCase: `internal/usecase/encrypt_decrypt.go`

**Generate encryption key:**
```bash
openssl rand -base64 32
```

**Ready for secure data handling!** 🔐

