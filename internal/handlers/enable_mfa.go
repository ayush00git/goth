package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

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
	
	// get userID from the metadata stored in UserContextKeys{}
	claims, ok := interceptors.GetClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	userID := claims.UserID

	switch mfa_type {
		case 1:
			// generate the totp_secret (random 32 bytes base-32 encoded string)
			totp_secret, err := helpers.GenerateTOTPSecret()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed generating totp_secret")
			}

			tags, err := g.db.Exec(
				ctx,
				`INSERT INTO mfa_secrets (user_id, totp_secret, mfa_type)
				VALUES ($1, $2, $3)
				ON CONFLICT (user_id) DO NOTHING`,
				userID,
				totp_secret,
				mfa_type,
			)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed inserting to database")
			}
			if tags.RowsAffected() == 0 {
				return nil, status.Error(codes.AlreadyExists, "mfa-type for this user already exists")
			}
			
			// generate the totp QR URL to be passed to the authenticator app
			QrURL := services.GenerateTOTPQRURL(claims.Email, totp_secret)

			return &goth.EnableMFAResponse{
				Success: true,
				TotpSecret: totp_secret,
				TotpQrUrl: QrURL,
				OtpExpiresIn: 0,
			}, nil
		case 2:
			// generate the otp
			otp, err := helpers.GenerateOTP()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed generating OTP")
			}

			// hash the otp
			hashedOTP, err := helpers.GenerateHash(otp)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed hashing otp")
			}

			// store to redis
			err = g.rdb.Set(
				ctx,
				"mfa:email_otp:" + userID,
				hashedOTP,
				5*time.Minute,
			).Err()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed storing to redis")
			}
	
			// store mfa_secrets to db
			tags, err := g.db.Exec(
				ctx,
				`INSERT INTO mfa_secrets (user_id, totp_secret, mfa_type)
				VALUES ($1, $2, $3)
				ON CONFLICT (user_id) DO NOTHING`,
				userID,
				"",
				mfa_type,
			)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed inserting mfa_secrets to database")
			}
			if tags.RowsAffected() == 0 {
				return nil, status.Error(codes.AlreadyExists, "mfa-type for this user already exists")
			}

			// create email subject
			mailSubject := "Email OTP MFA was enabled for your account"
			// create the email body
			mailBody := fmt.Sprintf(
				`	<h2>Verify MFA for your account</h2>
					<p>Your verification code is:</p>
					<h1>%s</h1>
					<p>This code expires in <b>5 minutes</b>.</p>
				`,
				otp,
			)
			
			// mail otp to the client
			email := claims.Email
			go func(email string) {
				if err := services.SendMail(claims.Email, mailSubject, mailBody); err != nil {
					log.Printf("Failed to send enable-mfa mail: %s", err)
				}
			}(email)

			return &goth.EnableMFAResponse{
				Success: true,
				TotpSecret: "",
				TotpQrUrl: "",
				OtpExpiresIn: 300,	// 5min
			}, nil
		default:
			return nil, status.Error(codes.InvalidArgument, "mfa_type argument not recognized")
	}
}
