package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/services"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoginReqBody
type RequestedUser struct {
	id				string
	email			string
	password		string
	email_verified	bool
	mfa_type		int16
}

func (g *GothServer) Login(ctx context.Context, req *goth.LoginRequest) (*goth.LoginResponse, error) {
	// inputs
	email := req.Email
	password := req.Password

	// validate the inputs
	if email == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required fields.")
	}

	// email lookup into db.
	var user RequestedUser
	err := g.db.QueryRow(
		ctx,
		`SELECT id, email, password_hash, email_verified, mfa_type
		FROM users
		WHERE email = $1`, email,
	).Scan(&user.id, &user.email, &user.password, &user.email_verified, &user.mfa_type)
	if err == pgx.ErrNoRows {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "database error")
	}

	// match the password
	err = bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "credentials do not match.")
	}

	// check if unverified
	if !user.email_verified {
		return nil, status.Error(codes.PermissionDenied, "unverified account")
	}

	// MFA on: no tokens yet. Return an mfa_session_token for VerifyMFA, and
	// email the code to email-OTP users. Tokens are issued by VerifyMFA.
	if mfaType := goth.MFAType(user.mfa_type); mfaType != goth.MFATypeNone {
		mfaSessionToken, err := services.GenerateMFASessionToken(user.id, user.email)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed generating mfa session token")
		}

		switch mfaType {
			case goth.MFATypeTotp:
				// nothing to send: the authenticator app generates the code,
				// and VerifyMFA checks it against mfa_secrets.totp_secret.

			case goth.MFATypeEmailOtp:
				// ResourceExhausted: a code was sent within the cooldown and is still valid.
				err := g.sendMFAEmailOTP(ctx, user.id, user.email)
				if err != nil && status.Code(err) != codes.ResourceExhausted {
					return nil, err
				}

			default:
				return nil, status.Error(codes.Internal, "unknown mfa_type for user")
		}

		return &goth.LoginResponse{
			MfaRequired: true,
			MfaSessionToken: mfaSessionToken,
			MfaType: mfaType,
		}, nil
	}

	// MFA off: issue tokens now
	// generate an access token on login
	accessToken, err := services.GenerateAccessToken(user.id, user.email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating an access token")
	}

	// generate a refresh token
	refreshToken, jti, err := services.GenerateRefreshToken(user.id, user.email, "go-client")
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating a refresh token")
	}

	// save jti to the db table
	_, err = g.db.Exec(
		ctx,
		`INSERT INTO REFRESH_TOKENS (user_id, jti, device_fingerprint)
		VALUES ($1, $2, $3)`,
		user.id,
		jti,
		"go-client",	// for testing
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to store refresh token to database")
	}

	return &goth.LoginResponse{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
		MfaRequired: false,
		MfaType: goth.MFATypeNone,
	}, nil
}
