package handlers

import (
	"context"
	"errors"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/interceptors"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdatePassword changes the authenticated user's password after checking the old one.
// All the user's refresh tokens are revoked, so other sessions must log in again.
// req.UserId is optional; if set it must match the access token.
func (g *GothServer) UpdatePassword(ctx context.Context, req *goth.UpdatePasswordRequest) (*goth.UpdatePasswordResponse, error) {
	claims, ok := interceptors.GetClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	userID := claims.UserID

	if req.UserId != "" && req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "cannot change another user's password")
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "old_password and new_password are required")
	}
	if req.OldPassword == req.NewPassword {
		return nil, status.Error(codes.InvalidArgument, "new password must be different from the old one")
	}

	// check the old password
	var passwordHash string
	err := g.db.QueryRow(
		ctx,
		`SELECT password_hash FROM users WHERE id = $1`,
		userID,
	).Scan(&passwordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed fetching user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.OldPassword)); err != nil {
		return nil, status.Error(codes.PermissionDenied, "old password is incorrect")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed hashing the password")
	}

	// update the password and revoke every session together
	tx, err := g.db.Begin(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed starting database transaction")
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2`,
		string(newHash),
		userID,
	)
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
		return nil, status.Error(codes.Internal, "failed committing password update")
	}

	return &goth.UpdatePasswordResponse{Success: true}, nil
}
