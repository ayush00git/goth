package helpers

import (
	"crypto/rand"
	"encoding/base32"
)

// GenerateTOTPSecret generates a random 32 bytes string
// base32 encode it and returns it as the totp_secret.
func GenerateTOTPSecret() (string, error) {
	secretBytes := make([]byte, 32)

	if _, err := rand.Read(secretBytes); err != nil {
		return "", err
	}

	encodedSecret := base32.StdEncoding.EncodeToString(secretBytes)
	return encodedSecret, nil
}
