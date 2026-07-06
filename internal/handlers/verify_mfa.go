package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/helpers"
	"github.com/ayush00git/goth/internal/services"
	"github.com/pquerna/otp/totp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Creds struct {
	MfaType		int32
	TotpSecret	string
}

// VerifyMFA completes the MFA of client based on its mfa_type and returns
// accesstoken and refresh token, authenticating the client.
func (g *GothServer) VerifyMFA(ctx context.Context, req *goth.VerifyMFARequest) (*goth.VerifyMFAResponse, error) {
	code := req.Code
	mfaSessionToken := req.MfaSessionToken

	claims, err := services.ValidateMFASessionToken(mfaSessionToken)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "complete password authentication first")
	}

	userID := claims.UserID
	email := claims.Email
	
	// read the mfa_secrets table
	var initials Creds
	err = g.db.QueryRow(
		ctx,
		`SELECT mfa_type, totp_secret
		FROM mfa_secrets
		WHERE user_id = $1`,
		userID,
	).Scan(&initials.MfaType, &initials.TotpSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed fetching mfa_type for user")
	}

	switch initials.MfaType {
		// TOTP based MFA
		case 1:
			// validate the client entered TOTP by generating it on server-side
			// via the same totp_secret.
			valid := totp.Validate(code, initials.TotpSecret)	// computes the check by the RPC 6238 algorithm
			if !valid {
				return nil, status.Error(codes.PermissionDenied, "invalid code")
			}

		case 2:
			// Email OTP based MFA
			// validate the hashedOTP stored in redis with the code client sends
			hashedCode, err := g.rdb.Get(
				ctx,
				"mfa:email_otp:" + userID,
			).Result()
			if err != nil {
				return nil, status.Error(codes.NotFound, "code may be expired")
			}

			err = helpers.ValidateHash(hashedCode, code)
			if err != nil {
				return nil, status.Error(codes.PermissionDenied, "invalid code")
			}

			// delete key from redis
			// if redis key deletion fails it could be ignored
			g.rdb.Del(ctx, "mfa:email_otp:" + userID)
		default:
    		return nil, status.Error(codes.InvalidArgument, "invalid mfa_type")
	}

	// update mfa_type in users table
	tags, err := g.db.Exec(
		ctx,
		`UPDATE users
		SET mfa_type = $1
		WHERE id = $2`,
		initials.MfaType,
		userID,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed updating users for mfa_type")
	}
	if tags.RowsAffected() == 0 {
		return nil, status.Error(codes.NotFound, "no user record updated")
	}

	// generate the access and refresh token
	accessToken, err := services.GenerateAccessToken(userID, email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating an access token")
	}

	refreshToken, _, err := services.GenerateRefreshToken(userID, email, "go-client")
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating refresh token")
	}

	return &goth.VerifyMFAResponse{
		Success: true,
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}
