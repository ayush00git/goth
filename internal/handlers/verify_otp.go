package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/helpers"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// consumeOTP checks code against the hashed OTP stored at otpKey.
// On success the OTP is deleted, so each code works once.
// After maxOTPAttempts wrong codes the OTP is deleted and a new one must be requested.
func (g *GothServer) consumeOTP(ctx context.Context, otpKey, code string) error {
	attemptsKey := "otp_attempts:" + otpKey

	hashedOTP, err := g.rdb.Get(ctx, otpKey).Result()
	if errors.Is(err, redis.Nil) {
		return status.Error(codes.NotFound, "code expired or not requested")
	}
	if err != nil {
		return status.Error(codes.Internal, "failed reading otp from redis")
	}

	if err := helpers.ValidateHash(hashedOTP, code); err != nil {
		attempts, err := g.rdb.Incr(ctx, attemptsKey).Result()
		if err != nil {
			return status.Error(codes.Internal, "failed counting otp attempts")
		}
		if attempts == 1 {
			g.rdb.Expire(ctx, attemptsKey, otpTTL)
		}
		if attempts >= maxOTPAttempts {
			g.rdb.Del(ctx, otpKey, attemptsKey)
			return status.Error(codes.ResourceExhausted, "too many wrong attempts, request a new code")
		}
		return status.Error(codes.PermissionDenied, "invalid code")
	}

	// Del returns how many keys it removed. 0 means a parallel request
	// already used this code, so only one request can win.
	deleted, err := g.rdb.Del(ctx, otpKey).Result()
	if err != nil {
		return status.Error(codes.Internal, "failed deleting otp from redis")
	}
	if deleted == 0 {
		return status.Error(codes.PermissionDenied, "code already used")
	}
	g.rdb.Del(ctx, attemptsKey)

	return nil
}

// VerifyOTP verifies an OTP sent by SendOTP for the given purpose.
//   - MFA: rejected, those codes belong to VerifyMFA and DisableMFA.
//   - Password reset: returns a one-time session_token for the password reset step.
//   - Email verify: marks the user's email as verified.
func (g *GothServer) VerifyOTP(ctx context.Context, req *goth.VerifyOTPRequest) (*goth.VerifyOTPResponse, error) {
	if req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}

	switch req.Purpose {
		case goth.OTPPurposeMfa:
			// MFA codes are checked by VerifyMFA (login) and DisableMFA. Verifying here
			// would delete the code before those rpcs could use it.
			return nil, status.Error(codes.InvalidArgument, "MFA codes are verified by VerifyMFA or DisableMFA")

		case goth.OTPPurposePasswordReset:
			email := req.Email
			if email == "" {
				return nil, status.Error(codes.InvalidArgument, "email is required")
			}

			if err := g.consumeOTP(ctx, "password_reset:email_otp:"+email, req.Code); err != nil {
				return nil, err
			}

			// random one-time token that proves the code was verified.
			// the password reset rpc will exchange it (and delete it) to set a new password.
			buf := make([]byte, 32)
			if _, err := rand.Read(buf); err != nil {
				return nil, status.Error(codes.Internal, "failed generating reset token")
			}
			resetToken := hex.EncodeToString(buf)

			err := g.rdb.Set(
				ctx,
				"password_reset:session:"+resetToken,
				email,
				passwordResetSessionTTL,
			).Err()
			if err != nil {
				return nil, status.Error(codes.Internal, "failed storing reset token")
			}

			return &goth.VerifyOTPResponse{
				Success: true,
				SessionToken: resetToken,
			}, nil

		case goth.OTPPurposeEmailVerify:
			email := req.Email
			if email == "" {
				return nil, status.Error(codes.InvalidArgument, "email is required")
			}

			if err := g.consumeOTP(ctx, "acc_verify:email_otp:"+email, req.Code); err != nil {
				return nil, err
			}

			tags, err := g.db.Exec(
				ctx,
				`UPDATE users
				SET email_verified = TRUE, updated_at = NOW()
				WHERE email = $1`,
				email,
			)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed updating email_verified")
			}
			if tags.RowsAffected() == 0 {
				return nil, status.Error(codes.NotFound, "user not found")
			}

			return &goth.VerifyOTPResponse{
				Success: true,
			}, nil

		default:
			return nil, status.Error(codes.InvalidArgument, "purpose code not defined.")
	}
}
