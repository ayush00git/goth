package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/helpers"
	"github.com/ayush00git/goth/internal/interceptors"
	"github.com/ayush00git/goth/internal/services"
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

	// userID := req.UserId		// UserId does comes from request body, but can't be trusted, fetch it from UserContextKeys{} instead
	claims, ok := interceptors.GetClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	userID := claims.UserID

	// start the databse transaction
	tx, err := g.db.Begin(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "error starting database transaction")
	}
	// register this into a new row in mfa_secrets table
	defer tx.Rollback(ctx)

	// generate the totp_secret (random 32 bytes base-32 encoded string)
	totp_secret, err := helpers.GenerateTOTPSecret()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating totp_secret")
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO mfa_secrets (user_id, totp_secret, mfa_type)
		VALUES ($1, $2, $3)`,
		userID,
		totp_secret,
		mfa_type,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed inserting to database")
	}
	
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, status.Error(codes.Internal, "failed committing the database transaction")
	}

	// generate the totp QR URL to be passed to the authenticator app
	QrURL := services.GenerateTOTPQRURL(claims.Email, totp_secret)

	return &goth.EnableMFAResponse{
		Success: true,
		TotpSecret: totp_secret,
		TotpQrUrl: QrURL,
		OtpExpiresIn: 0,
	}, nil
}
