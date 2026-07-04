package helpers

import "golang.org/x/crypto/bcrypt"

// Hash hashes a input string using the bcrypt library
// and returns a hashed string.
func GenerateHash(input string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(input), 10)
	if err != nil {
		return "", nil
	}

	hashedString := string(hashedBytes)
	return hashedString, err
}
