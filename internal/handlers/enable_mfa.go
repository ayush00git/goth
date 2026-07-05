package handlers

import (
	"context"
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

			_, err = g.db.Exec(
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
			// mail otp to the client
			go func() {
				if err := services.SendMail(claims.Email, "Email OTP MFA was enabled for your account"); err != nil {
					log.Printf("Failed to send enable-mfa mail: %s", err)
				}
			}()

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
