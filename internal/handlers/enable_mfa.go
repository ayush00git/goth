package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EnableMFA allows client to enable multi-factor authentication for their
// accounts. It accepts the mfa_type and activates the same for the user.
func (g *GothServer) EnableMFA(ctx context.Context, req *goth.EnableMFARequest) (*goth.EnableMFAResponse, error) {
	// client will send MFA type to be enabled for the user
	// mfa_type = 0 : None
	//			= 1 : TOTP
	// 			= 2 : EmailOTP
	mfa_type := req.MfaType
	userID := req.UserId

	tx, err := g.db.Begin(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "error starting database transaction")
	}
	// register this into a new row in mfa_secrets table
	defer tx.Rollback(ctx)
	_, err = tx.Exec(
		ctx,
		`INSERT INTO mfa_secrets (user_id, totp_secret, mfa_type)
		VALUES ($1, $2, $3)`,
		userID,
		"totp_secret",
		mfa_type,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed inserting to database")
	}
	
	// register the same mfa_type in the users table
	tag, err := tx.Exec(
		ctx,
		`UPDATE users
		SET mfa_type = $1
		WHERE id = $2`,
		mfa_type,
		userID,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed updating mfa_type to users")
	}
	if tag.RowsAffected() == 0 {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, status.Error(codes.Internal, "failed committing the database transaction")
	}

	return &goth.EnableMFAResponse{
		Success: true,
		TotpSecret: 
	}, nil
}
