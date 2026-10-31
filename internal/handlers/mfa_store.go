package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// users.mfa_type is the source of truth for "is MFA on for this user".
// mfa_secrets only holds the secret. Both are always written together.

// saveMFASecret stores the user's MFA secret and turns MFA on in users.mfa_type.
func (g *GothServer) saveMFASecret(ctx context.Context, userID, totpSecret string, mfaType goth.MFAType) error {
	tx, err := g.db.Begin(ctx)
	if err != nil {
		return status.Error(codes.Internal, "failed starting database transaction")
	}
	defer tx.Rollback(ctx)

	tags, err := tx.Exec(
		ctx,
		`INSERT INTO mfa_secrets (user_id, totp_secret, mfa_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO NOTHING`,
		userID,
		totpSecret,
		mfaType,
	)
	if err != nil {
		return status.Error(codes.Internal, "failed inserting mfa_secrets to database")
	}
	if tags.RowsAffected() == 0 {
		return status.Error(codes.AlreadyExists, "mfa-type for this user already exists")
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE users
		SET mfa_type = $1, updated_at = NOW()
		WHERE id = $2`,
		mfaType,
		userID,
	)
	if err != nil {
		return status.Error(codes.Internal, "failed updating mfa_type for user")
	}

	if err := tx.Commit(ctx); err != nil {
		return status.Error(codes.Internal, "failed committing mfa_secrets")
	}
	return nil
}

// removeMFASecret deletes the user's MFA secret and turns MFA off in users.mfa_type.
func (g *GothServer) removeMFASecret(ctx context.Context, userID string) error {
	tx, err := g.db.Begin(ctx)
	if err != nil {
		return status.Error(codes.Internal, "failed starting database transaction")
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`DELETE FROM mfa_secrets
		WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return status.Error(codes.Internal, "failed to delete the mfa_secrets")
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE users
		SET mfa_type = $1, updated_at = NOW()
		WHERE id = $2`,
		goth.MFATypeNone,
		userID,
	)
	if err != nil {
		return status.Error(codes.Internal, "failed resetting mfa_type for user")
	}

	if err := tx.Commit(ctx); err != nil {
		return status.Error(codes.Internal, "failed committing mfa removal")
	}
	return nil
}
