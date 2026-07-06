package services

import (
	"errors"
	"time"

	"github.com/ayush00git/goth/internal/helpers"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Helper function which generates a refresh token which helps re-generating the
// access tokens, this token is stored in a database and revoked after a specific time.
func GenerateRefreshToken(userId, email, fingerprint string) (tokenString string, jti string, err error) {
	secretKey := helpers.GetEnvVar("REFRESH_TOKEN_SECRET")
	jti = uuid.NewString()
	// build the claims object.
	claims := Claims{
		UserID: userId,
		Email: email,
		DeviceFingerprint: fingerprint,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:			jti,
			Subject:	userId,
			IssuedAt: 	jwt.NewNumericDate(time.Now()),
			ExpiresAt: 	jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),	// expiration time - 24 hours
		},
	}

	// building the token object with "NewWithClaims"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err = token.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	return tokenString, jti, nil
}

// Helper function which validates the refresh token, just required using
// logout to revoke the current stored refresh token.
func ValidateRefreshToken(refreshTokenString string) (*Claims, error) {
	secretKey := helpers.GetEnvVar("REFRESH_TOKEN_SECRET")
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &Claims{},
		func(refreshToken *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
	)
	if err != nil {
		return nil, err
	}	

	claims, ok := refreshToken.Claims.(*Claims)
	if !ok || !refreshToken.Valid {
		return nil, errors.New("Invalid token")
	}

	return claims, nil
}
