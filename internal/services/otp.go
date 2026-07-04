package services

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

func GenerateOTP() (string, error) {
	randomBytes := make([]byte, 4)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	otp := binary.BigEndian.Uint32(randomBytes) % 1000000	// truncate to six decimal places
	return fmt.Sprintf("%06d", otp), nil
}
