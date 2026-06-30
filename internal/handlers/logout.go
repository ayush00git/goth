package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GothServer) Logout(ctx context.Context, req *goth.LogoutRequest) (*goth.LogoutResponse, error) {
	claims, err := services.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unverified refresh token")
	}
	// revoke the refresh token
	tag, err := g.db.Exec(
		ctx,
		`UPDATE REFRESH_TOKENS
		SET revoked = TRUE
		WHERE jti = $1`,
		claims.ID,		// claims.ID = jti
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed revoking refresh token")
	}
	if tag.RowsAffected() == 0 {
    	return nil, status.Error(codes.NotFound, "session not found")
	}

	return &goth.LogoutResponse{Success: true}, nil
}
