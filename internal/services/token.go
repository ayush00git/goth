package services

import "github.com/golang-jwt/jwt/v5"

// jwt payload claims
type Claims struct {
	UserID				string		`json:"user_id"`
	Email				string		`json:"email"`
	DeviceFingerprint	string		`json:"device_fingerprint,omitempty"`
	jwt.RegisteredClaims
}
