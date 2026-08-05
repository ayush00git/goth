package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/helpers"
	"github.com/ayush00git/goth/internal/interceptors"
	"github.com/ayush00git/goth/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GothServer) SendOTP(ctx context.Context, req *goth.SendOTPRequest) (*goth.SendOTPResponse, error) {
	// OTPPurpose is a enum type with specified values for each type.
	// 0 - MfaType (authenticated)
	// 1 - Password reset (user should be authenticated validate email via access token)
	// 2 - Email Verification (can accept email from the user in this case).
	purpose := req.Purpose

	switch purpose {
		case 0:
			// get the email of the authenticated user from the access token.
			claims, ok := interceptors.GetClaims(ctx)
			if !ok {
				return nil, status.Error(codes.Unauthenticated, "failed getting claims")
			}

			email := claims.Email
			userID := claims.UserID

			otp, err := helpers.GenerateOTP()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed generating OTP")
			}

			hashedOTP, err := helpers.GenerateHash(otp)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed hashing the otp")
			}

			// store OTP to redis.
			err = g.rdb.Set(
				ctx,
				"mfa:email_otp:" + userID,
				hashedOTP,
				5*time.Minute,
			).Err()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed storing otp to redis")
			}

			mailBody := fmt.Sprintf(
				`	<h2>Verify MFA for your account</h2>
					<p>Your verification code is:</p>
					<h1>%s</h1>
					<p>This code expires in <b>5 minutes</b>.</p>
				`,
				otp,
			)

			err = services.SendMail(email, "Verify MFA for your account", mailBody)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed sending email to user")
			}
		
		case 1:
			// get the email of the authenticated user from the access token.
			claims, ok := interceptors.GetClaims(ctx)
			if !ok {
				return nil, status.Error(codes.Internal, "failed getting claims")
			}

			email := claims.Email

	}
}
