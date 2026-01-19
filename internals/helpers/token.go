package helpers

import (
	"time"
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/golang-jwt/jwt/v5"
)

// Blueprint of a token
type Claims struct {
	UserID		string		`json:"user_id"`
	Email		string		`json:"email"`
	jwt.RegisteredClaims
}

// TO retrieve the secret key
func GetSecretKey() []byte {
	_ = godotenv.Load()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("fallback-error-only-for-devs")
	}
	return []byte(secret)
}

// Generating a token
func GenerateToken(userId, email string) (string, error) {
	claims := Claims{
		UserID: userId,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()), 
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetSecretKey())
}

// Verifying a token
func VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return GetSecretKey(), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("Invalid token")
}
