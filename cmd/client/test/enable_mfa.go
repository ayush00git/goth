package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"google.golang.org/grpc/metadata"
)

func TestEnableMFA(c goth.GothServiceClient, accessToken string) {
    md := metadata.Pairs(
        "authorization",
        "Bearer "+accessToken,
    )

    ctx := metadata.NewOutgoingContext(context.Background(), md)

	resp, err := c.EnableMFA(ctx, &goth.EnableMFARequest{
		MfaType: 2,
	})
	if err != nil {
		panic((err))
	}

	fmt.Print("MFA Enabled\n")
	fmt.Printf("success: %t\n", resp.Success)
	fmt.Printf("totp_secret: %s\n", resp.TotpSecret)
	fmt.Printf("totp_QrUrl: %s\n", resp.TotpQrUrl)
	fmt.Printf("otp_expires_in: %d\n", resp.OtpExpiresIn)
}
