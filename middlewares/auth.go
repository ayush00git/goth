package middlewares

import (
	"net/http"
	"strings"

	"goth/helpers"

	"github.com/gin-gonic/gin"
)

// context keys
const (
	UserIDKey = "userID"
	RoleIDKey = "role"
	UserNameKey = "userName"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		// try to get tokenString from cookies
		cookie, err := c.Cookie("token")
		if err == nil {
			tokenString = cookie
		}

		// try to find in headers (authorization header)
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 {
					tokenString = parts[1]
				}
			}
		}

		// no token found
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access is not allowed!"})
			c.Abort()
			return
		}

		// verify the token
		claims, err := helpers.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not recognized!"})
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.ID)
		c.Set(RoleIDKey, claims.Role)
		c.Set(UserNameKey, claims.UserName)

		c.Next()
	}
}
