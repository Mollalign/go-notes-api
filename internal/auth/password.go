package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory      uint32 = 19 * 1024
	argonIterations  uint32 = 2
	argonParallelism uint8  = 1
	saltLength              = 16
	keyLength               = 32
)

var ErrInvalidPasswordHash = errors.New("invalid password hash")

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonIterations,
		argonMemory,
		argonParallelism,
		keyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonIterations,
		argonParallelism,
		encodedSalt,
		encodedHash,
	), nil
}

func VerifyPassword(encodedHash, password string) (bool, error) {
	parts := strings.Split(encodedHash, "$")

	if len(parts) != 5 ||
		parts[0] != "argon2id" ||
		parts[1] != "v=19" ||
		parts[2] != "m=19456,t=2,p=1" {
		return false, ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(salt) != saltLength {
		return false, ErrInvalidPasswordHash
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(expectedHash) != keyLength {
		return false, ErrInvalidPasswordHash
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		argonIterations,
		argonMemory,
		argonParallelism,
		keyLength,
	)

	return subtle.ConstantTimeCompare(
		actualHash,
		expectedHash,
	) == 1, nil
}
