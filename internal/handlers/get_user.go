package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/interceptors"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUser returns the profile of the authenticated user.
// req.UserId is optional; if set it must match the access token, so a user
// can never read someone else's profile by changing the id.
func (g *GothServer) GetUser(ctx context.Context, req *goth.GetUserRequest) (*goth.GetUserResponse, error) {
	claims, ok := interceptors.GetClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.UserId != "" && req.UserId != claims.UserID {
		return nil, status.Error(codes.PermissionDenied, "cannot read another user's profile")
	}

	var (
		user      goth.GetUserResponse
		mfaType   int16
		createdAt time.Time
	)
	err := g.db.QueryRow(
		ctx,
		`SELECT id, email, full_name,
			COALESCE(email_verified, FALSE),
			COALESCE(mfa_type, 0),
			created_at
		FROM users
		WHERE id = $1`,
		claims.UserID,
	).Scan(&user.UserId, &user.Email, &user.FullName, &user.EmailVerified, &mfaType, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed fetching user")
	}

	return &goth.GetUserResponse{
		UserId: user.UserId,
		Email: user.Email,
		FullName: user.FullName,
		EmailVerified: user.EmailVerified,
		MfaType: goth.MFAType(mfaType),
		CreatedAt: createdAt.Unix(),
	}, nil
}
