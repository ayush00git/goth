package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwt payload claims
type Claims struct {
	UserID				string		`json:"user_id"`
	Email				string		`json:"email"`
	DeviceFingerprint	string		`json:"device_fingerprint,omitempty"`
	jwt.RegisteredClaims
}

// Helper function which generates an access token which helps identifying
// authorized users at every protected handlers/rpc services.
func GenerateAccessToken (userId, email string) (string, error) {
	// define the claims/payload including the registered claims.
	claims := Claims{
		UserID: userId,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),	// expiration time - 15 min
		},
	}
	// "token" object with "NewWithClaims" defining the signing
	// method and claims contained in the object.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign the token string.
	tokenString, err := token.SignedString([]byte("jwuweufwfweif"))		// move it to environment variabls after
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Helper function which validates the user's authorization via his access token
// which allows user's access to protected rpc services.
func ValidateAccessToken(tokenString string) (*Claims, error) {
	secretKey := []byte("jwuweufwfweif")	// move to env variables after
	// parse the jwt token string.
	token, err := jwt.ParseWithClaims(tokenString, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	
	// validate the token.
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("Invalid Token")
}

// Helper function which generates a refresh token which helps re-generating the
// access tokens, this token is stored in a database and revoked after a specific time.
func GenerateRefreshToken(userId, email, fingerprint string) (string, error) {
	// build the claims object.
	claims := Claims{
		UserID: userId,
		Email: email,
		DeviceFingerprint: fingerprint,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),		// expiration time - 24 hours
		},
	}

	// building the token object with "NewWithClaims"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("euiy348bcues"))	// move it to env vars later.
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
