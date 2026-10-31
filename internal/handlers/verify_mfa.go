package handlers

import (
	"context"
	"errors"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/services"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Creds struct {
	MfaType		int32
	TotpSecret	*string
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
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.FailedPrecondition, "MFA is not enabled for this user")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed fetching mfa_type for user")
	}

	switch initials.MfaType {
		// TOTP based MFA
		case 1:
			// validate the TOTP with a limit on wrong attempts
			if err := g.checkTOTP(ctx, userID, *initials.TotpSecret, code); err != nil {
				return nil, err
			}

		case 2:
			// Email OTP based MFA
			// check the code, delete it once used, and burn it after too many wrong tries
			if err := g.consumeOTP(ctx, "mfa:email_otp:"+userID, code); err != nil {
				return nil, err
			}
		default:
    		return nil, status.Error(codes.InvalidArgument, "invalid mfa_type")
	}

	// generate the access and refresh token
	accessToken, err := services.GenerateAccessToken(userID, email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating an access token")
	}

	refreshToken, jti, err := services.GenerateRefreshToken(userID, email, "go-client")
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating refresh token")
	}

	// save jti so RefreshToken and Logout can find this session
	_, err = g.db.Exec(
		ctx,
		`INSERT INTO refresh_tokens (user_id, jti, device_fingerprint)
		VALUES ($1, $2, $3)`,
		userID,
		jti,
		"go-client",	// for testing
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to store refresh token to database")
	}

	return &goth.VerifyMFAResponse{
		Success: true,
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}
