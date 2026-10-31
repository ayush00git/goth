package handlers

import (
	"context"
	"errors"
	"log"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/interceptors"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GothServer) DisableMFA(ctx context.Context, req *goth.DisableMFARequest) (*goth.DisableMFAResponse, error) {
	code := req.Code
	// access tokens would be authenticated by the auth interceptor
	claims, ok := interceptors.GetClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed parsing the claims")
	}

	userID := claims.UserID

	var initials Creds
	err := g.db.QueryRow(
		ctx,
		`SELECT mfa_type, totp_secret
		FROM mfa_secrets
		WHERE user_id = $1`,
		userID,
	).Scan(&initials.MfaType, &initials.TotpSecret)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.FailedPrecondition, "MFA is not enabled for this user")
	}
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, status.Error(codes.Internal, "failed to read the database")
	}

	switch initials.MfaType {
		case 1:
			// validate the TOTP with a limit on wrong attempts
			if err := g.checkTOTP(ctx, userID, *initials.TotpSecret, code); err != nil {
				return nil, err
			}
			
			// delete mfa_secrets and turn MFA off in users.mfa_type, together
			tx, err := g.db.Begin(ctx)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed starting database transaction")
			}
			defer tx.Rollback(ctx)

			_, err = tx.Exec(
				ctx,
				`DELETE FROM mfa_secrets
				WHERE user_id = $1`,
				userID,
			)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to delete the mfa_secrets")
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
				return nil, status.Error(codes.Internal, "failed resetting mfa_type for user")
			}

			if err := tx.Commit(ctx); err != nil {
				return nil, status.Error(codes.Internal, "failed committing mfa removal")
			}
		
		case 2:
			// check the code, delete it once used, and burn it after too many wrong tries
			if err := g.consumeOTP(ctx, "mfa:email_otp:"+userID, code); err != nil {
				return nil, err
			}

			// delete mfa_secrets and turn MFA off in users.mfa_type, together
			tx, err := g.db.Begin(ctx)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed starting database transaction")
			}
			defer tx.Rollback(ctx)

			_, err = tx.Exec(
				ctx,
				`DELETE FROM mfa_secrets
				WHERE user_id = $1`,
				userID,
			)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to delete the mfa_secrets")
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
				return nil, status.Error(codes.Internal, "failed resetting mfa_type for user")
			}

			if err := tx.Commit(ctx); err != nil {
				return nil, status.Error(codes.Internal, "failed committing mfa removal")
			}

		default:
			return nil, status.Error(codes.InvalidArgument, "invalid mfa_type")
		}

	return &goth.DisableMFAResponse{
		Success: true,
	}, nil
}
