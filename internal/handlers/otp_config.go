package handlers

import "time"

// OTP timings and limits shared by SendOTP, VerifyOTP and EnableMFA.
// Change them here and every handler (and the email text) follows.
const (
	// otpTTL is how long an emailed OTP stays valid.
	otpTTL = 5 * time.Minute

	// otpResendCooldown is the minimum gap between two OTP emails for the same target.
	otpResendCooldown = 60 * time.Second

	// maxOTPAttempts is how many wrong codes are allowed before the OTP is burned.
	maxOTPAttempts = 5

	// passwordResetSessionTTL is how long the reset token from VerifyOTP stays usable.
	passwordResetSessionTTL = 10 * time.Minute
)
