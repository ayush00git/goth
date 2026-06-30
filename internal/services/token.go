package services

import (
	"errors"
	"time"

	"github.com/ayush00git/goth/internal/helpers"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
func GenerateAccessToken(userId, email string) (string, error) {
	secretKey := helpers.GetEnvVar("ACCESS_TOKEN_SECRET")
	// define the claims/payload including the registered claims.
	claims := Claims{
		UserID: userId,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:	userId,
			IssuedAt: 	jwt.NewNumericDate(time.Now()),
			ExpiresAt: 	jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),	// expiration time - 15 min
		},
	}
	// "token" object with "NewWithClaims" defining the signing
	// method and claims contained in the object.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign the token string.
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Helper function which validates the user's authorization via his access token
// which allows user's access to protected rpc services.
func ValidateAccessToken(tokenString string) (*Claims, error) {
	secretKey := helpers.GetEnvVar("ACCESS_TOKEN_SECRET")
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
	return nil, err
}

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

func RefreshAccessToken(refreshToken string) (string, error) {
	refreshSecret := helpers.GetEnvVar("REFRESH_TOKEN_SECRET")
	accessSecret := helpers.GetEnvVar("ACCESS_TOKEN_SECRET")
	// validate the refresh token and then pass on those
	// values to generate a new access token
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
		return refreshSecret, nil
		},
	)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*Claims)
	if !token.Valid || !ok {
		return "", errors.New("Invalid token")
	}

	newClaims := Claims{
		UserID: claims.UserID,
		Email:	claims.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: 	claims.UserID,
			IssuedAt:	jwt.NewNumericDate(time.Now()),
			ExpiresAt: 	jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)

	accessTokenString, err := accessToken.SignedString([]byte(accessSecret))
	if err != nil {
		return "", err
	}
	return accessTokenString, nil
}
