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
	// 0 - MfaType (authenticated user is required).
	// 1 - Password reset (user could be auth/unauthenticated).
	// 2 - Email Verification (user could be auth/unauthenticated).
	

	switch req.Purpose {
		case goth.OTPPurposeMfa:
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

			return &goth.SendOTPResponse{
				Success: true,
				ExpiresInSeconds: 300,
			}, nil
		
		case goth.OTPPurposePasswordReset:

			email := req.Email
			
			// generate + hash the otp and then store to redis.
			otp, err := helpers.GenerateOTP()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed generating OTP")
			}

			hashedOTP, err := helpers.GenerateHash(otp)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed hashing the otp")
			}

			// store to redis.
			err = g.rdb.Set(
				ctx,
				"password_reset:email_otp" + email,
				hashedOTP,
				5*time.Minute,
			).Err()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed storing to redis")
			}

			mailBody := fmt.Sprintf(
				`	<h2>Reset Password for your account</h2>
					<p>Your verification code is:</p>
					<h1>%s</h1>
					<p>This code expires in <b>5 minutes</b>.</p>
				`,
				otp,
			)

			err = services.SendMail(email, "Reset Password for your account", mailBody)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed sending email to user")
			}

			return &goth.SendOTPResponse{
				Success: true,
				ExpiresInSeconds: 300,
			}, nil

		case goth.OTPPurposeEmailVerify:
			// get the email once again.
			email := req.Email

			otp, err := helpers.GenerateOTP()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed generating OTP")
			}

			hashedOTP, err := helpers.GenerateHash(otp)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed hashing the otp")
			}

			// store to redis.
			err = g.rdb.Set(
				ctx,
				"acc_verify:email_otp" + email,
				hashedOTP,
				5*time.Minute,
			).Err()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed storing to redis")
			}

			mailBody := fmt.Sprintf(
				`	<h2>Verify your account.</h2>
					<p>Your verification code is:</p>
					<h1>%s</h1>
					<p>This code expires in <b>5 minutes</b>.</p>
				`,
				otp,
			)

			err = services.SendMail(email, "Verify your email", mailBody)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed sending email to user")
			}

			return &goth.SendOTPResponse{
				Success: true,
				ExpiresInSeconds: 300,
			}, nil

		default:
			return nil, status.Error(codes.InvalidArgument, "purpose code not defined.")
	}
}
