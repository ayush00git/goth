package middlewares

import (
	"context"
	"net/http"
	"strings"

	"goth/helpers"
)

type ContextKey string
const (
	UserIDKey	ContextKey = "userId"
	RoleIDKey	ContextKey = "role"
	UserNameKey	ContextKey = "userName"
)

func AuthMiddleware (next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := ""

		cookie, err := r.Cookie("token")
		if err == nil {
			tokenString = cookie.Value
		}

		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}

		claims, err := helpers.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "Token not recogonized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)		// req.Context(), key, value
		ctx = context.WithValue(ctx, RoleIDKey, claims.Role)
		ctx = context.WithValue(ctx, UserNameKey, claims.UserName)
		
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

