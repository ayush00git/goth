package services

import (
	"errors"
	"time"

	"github.com/ayush00git/goth/internal/helpers"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateMFASessionToken generates an MFASessionToken that is returned to
// client after his password login to be authorized for the mfa auth.
func GenerateMFASessionToken(userID, email string) (string, error) {
	secretKey := helpers.GetEnvVar("MFA_SESSION_TOKEN_SECRET")

	claims := &Claims{
		UserID: userID,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),	// 10 minutes
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateMFASessionToken validates a MFASessionToken and helps the corresponding
// rpc method to know the mfa_type of a user.
func ValidateMFASessionToken(tokenString string) (*Claims, error) {
	secretKey := helpers.GetEnvVar("MFA_SESSION_TOKEN_SECRET")
	
	token, err := jwt.ParseWithClaims(
		tokenString,
		Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid mfa session token")
	}

	return claims, nil
}
