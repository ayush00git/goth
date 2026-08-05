package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/helpers"
	"github.com/ayush00git/goth/internal/interceptors"
	"github.com/pquerna/otp/totp"
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
		WHERE id = $1`,
		userID,
	).Scan(&initials.MfaType, &initials.TotpSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to read the database")
	}

	switch initials.MfaType {
		case 1:
			// validate the client entered TOTP by generating it on server-side
			// via the same totp_secret.
			valid := totp.Validate(code, *initials.TotpSecret)	// computes the check by the RPC 6238 algorithm
			if !valid {
				return nil, status.Error(codes.PermissionDenied, "invalid code")
			}
			
			// Delete mfa_secrets for this user
			_, err := g.db.Exec(
				ctx,
				`DELETE FROM mfa_secrets
				WHERE user_id = $1`,
				userID,
			)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to delete the mfa_secrets")
			}
		
		case 2:
			// match the otp with the one stored in redis.
			hashCode, err := g.rdb.Get(
				ctx,
				"mfa:email_otp:" + userID,
			).Result()
			if err != nil {
				return nil, status.Error(codes.NotFound, "code may be expired")
			}

			err = helpers.ValidateHash(hashCode, code)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, "invalid code")
			}

			// Delete mfa_secrets for this user
			_, err = g.db.Exec(
				ctx,
				`DELETE FROM mfa_secrets
				WHERE user_id = $1`,
				userID,
			)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to delete the mfa_secrets")
			}

		default:
			return nil, status.Error(codes.InvalidArgument, "invalid mfa_type")
		}

	return &goth.DisableMFAResponse{
		Success: true,
	}, nil
}
