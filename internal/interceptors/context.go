package interceptors

import (
	"context"

	"github.com/ayush00git/goth/internal/services"
)

// to avoid plain strings storing the context keys
type UserContextKeys struct {}

// helper function to use instead of asserting type
// every time for UserContextKeys{}
func GetClaims(ctx context.Context) (*services.Claims, bool) {
	claims, ok := ctx.Value(UserContextKeys{}).(*services.Claims)
	return claims, ok
}
