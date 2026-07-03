package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
)

func TestEnableMFA(c goth.GothServiceClient) {
	resp, err := c.EnableMFA(context.Background(), &goth.EnableMFARequest{
		MfaType: 1,
	})
	if err != nil {
		panic((err))
	}

	fmt.Print("MFA Enabled\n")
	fmt.Printf("totp_secret: %s", resp.TotpSecret)
	fmt.Printf("totp_QrUrl: %s", resp.TotpQrUrl)
	// otp expiry for `mfa_type = 2`
}
