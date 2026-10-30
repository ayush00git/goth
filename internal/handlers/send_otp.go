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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// otpResendCooldown is the minimum gap between two OTP emails for the same target.
const otpResendCooldown = 60 * time.Second

// mfaIdentity returns the user an MFA OTP should go to. It accepts either:
//   - an access token (claims set by AuthInterceptor), for enabling/disabling MFA, or
//   - an "x-mfa-session-token" metadata header, for the login step, where the
//     user has passed the password check but has no access token yet.
func mfaIdentity(ctx context.Context) (userID, email string, err error) {
	if claims, ok := interceptors.GetClaims(ctx); ok {
		return claims.UserID, claims.Email, nil
	}

	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get("x-mfa-session-token")
	if len(values) == 0 {			// missing password check stage.
		return "", "", status.Error(codes.Unauthenticated, "access token or mfa session token required")
	}

	claims, err := services.ValidateMFASessionToken(values[0])
	if err != nil {
		return "", "", status.Error(codes.Unauthenticated, "invalid mfa session token")
	}

	return claims.UserID, claims.Email, nil
}

// otpCooldown blocks repeated OTP emails to the same target within otpResendCooldown.
func (g *GothServer) otpCooldown(ctx context.Context, target string) error {
	ok, err := g.rdb.SetNX(ctx, "otp_cooldown:"+target, 1, otpResendCooldown).Result()
	if err != nil {
		return status.Error(codes.Internal, "failed checking otp cooldown")
	}
	if !ok {
		return status.Error(codes.ResourceExhausted, "otp already sent, try again later")
	}
	return nil
}

func (g *GothServer) SendOTP(ctx context.Context, req *goth.SendOTPRequest) (*goth.SendOTPResponse, error) {
	// OTPPurpose is a enum type with specified values for each type.
	// 0 - MfaType (access token, or x-mfa-session-token metadata during login).
	// 1 - Password reset (user could be auth/unauthenticated).
	// 2 - Email Verification (user could be auth/unauthenticated).
	

	switch req.Purpose {
		case goth.OTPPurposeMfa:
			// never trust req.Email here: the email always comes from a signed token.
			userID, email, err := mfaIdentity(ctx)		// even if user has passed an email through the request body, this function takes care of it.
			if err != nil {
				return nil, err
			}

			if err := g.otpCooldown(ctx, "mfa:"+userID); err != nil {
				return nil, err
			}

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

			if err := g.otpCooldown(ctx, "password_reset:"+email); err != nil {
				return nil, err
			}
			
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
				"password_reset:email_otp:" + email,
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

			// check if the user's account is already verified.
			var isVerified bool
			err := g.db.QueryRow(
				ctx,
				`SELECT email_verified FROM users
				WHERE email = $1`,
				email,
			).Scan(&isVerified)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed sending email to user")
			}
			if isVerified == true {
				return nil, status.Error(codes.AlreadyExists, "your account is already verified")
			}

			if err := g.otpCooldown(ctx, "acc_verify:"+email); err != nil {
				return nil, err
			}

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
				"acc_verify:email_otp:" + email,
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
