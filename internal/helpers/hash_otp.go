package helpers

import "golang.org/x/crypto/bcrypt"

// GenerateHash hashes a input string using the bcrypt library
// and returns a hashed string.
func GenerateHash(input string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(input), 10)
	if err != nil {
		return "", nil
	}

	hashedString := string(hashedBytes)
	return hashedString, err
}

// ValidateHash Compares the hash with its actual string version
// and returns success.
func ValidateHash(hashedCode, code string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(code))
	if err != nil {
		return err
	}
	return nil
}
