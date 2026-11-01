package handlers

import (
	"context"
	"errors"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ResetPassword sets a new password for a user who forgot theirs.
// The session_token comes from VerifyOTP (purpose PASSWORD_RESET) and works once.
// All the user's refresh tokens are revoked, so every session must log in again.
func (g *GothServer) ResetPassword(ctx context.Context, req *goth.ResetPasswordRequest) (*goth.ResetPasswordResponse, error) {
	// validate before using the token, so a bad request doesn't burn it
	if req.SessionToken == "" || req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "session_token and new_password are required")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed hashing the password")
	}

	// GETDEL reads and deletes in one step, so the token can't be used twice
	email, err := g.rdb.GetDel(ctx, "password_reset:session:"+req.SessionToken).Result()
	if errors.Is(err, redis.Nil) {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired reset session")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed reading reset session")
	}

	// update the password and revoke every session together
	tx, err := g.db.Begin(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed starting database transaction")
	}
	defer tx.Rollback(ctx)

	var userID string
	err = tx.QueryRow(
		ctx,
		`UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE email = $2
		RETURNING id`,
		string(newHash),
		email,
	).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed updating password")
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE user_id = $1 AND revoked = FALSE`,
		userID,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed revoking sessions")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, status.Error(codes.Internal, "failed committing password reset")
	}

	return &goth.ResetPasswordResponse{Success: true}, nil
}
