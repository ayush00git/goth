package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// rpc handler which implements the handler logic for rpc method RefreshToken, accepts a
// refresh token and rotates the refresh token and returns an access token
func (g *GothServer) RefreshToken(ctx context.Context, req *goth.RefreshTokenRequest) (*goth.RefreshTokenResponse, error) {
	refreshTokenString := req.RefreshToken
	// parse the incoming the refresh token string
	claims, err := services.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated access")
	}

	// revoke the old refresh token
	oldJti := claims.ID		// registered claims `ID` is identification for the jwt

	// rotate the refresh token for the user
	// start the query transaction
	tx, err := g.db.Begin(ctx)
	if err != nil{
		return nil, status.Error(codes.Internal, "failed starting database transaction")
	}
	defer tx.Rollback(ctx)
	// revoke the old refresh token
	tag, err := tx.Exec(
		ctx,
		`UPDATE refresh_tokens
		SET revoked = true
		WHERE jti = $1 AND revoked = false`, // revoked = false handles the race conditions
		oldJti,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed revoking token")
	}
	if tag.RowsAffected() == 0 {
		return nil, status.Error(codes.Unauthenticated, "refresh token is invalid or just got revoked")
	}

	// generate a new refresh token and create a new row using this newJti
	refreshTokenString, newJti, err := services.GenerateRefreshToken(claims.UserID, claims.Email, claims.DeviceFingerprint)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating refresh token")
	}

	// insert a new refresh token into database for this user
	_, err = tx.Exec(
		ctx,
		`INSERT INTO refresh_tokens (user_id, jti, device_fingerprint)
		VALUES ($1, $2, $3)
		`,
		claims.UserID,
		newJti,
		claims.DeviceFingerprint,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed inserting to database")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, status.Error(codes.Internal, "failed updating database")
	}
	
	// generate an access token unconditionally
	accessTokenString, err := services.GenerateAccessToken(claims.UserID, claims.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating access token")
	}

	return &goth.RefreshTokenResponse{
		AccessToken: accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
