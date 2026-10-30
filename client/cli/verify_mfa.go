package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
)

var verifyMFACmd = &cobra.Command{
	Use:   "verify-mfa",
	Short: "Complete MFA with the session token from login and a TOTP/email code",
	RunE: func(cmd *cobra.Command, args []string) error {
		mfaSessionToken, _ := cmd.Flags().GetString("mfa-session-token")
		code, _ := cmd.Flags().GetString("code")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		// no access token: VerifyMFA is public and checks the session token itself.
		resp, err := client.VerifyMFA(context.Background(), &goth.VerifyMFARequest{
			MfaSessionToken: mfaSessionToken,
			Code:            code,
		})
		if err != nil {
			return err
		}

		fmt.Println("MFA Verified")
		fmt.Printf("Success:       %t\n", resp.Success)
		fmt.Printf("Access Token:  %s\n", resp.AccessToken)
		fmt.Printf("Refresh Token: %s\n", resp.RefreshToken)
		return nil
	},
}

func init() {
	verifyMFACmd.Flags().String("mfa-session-token", "", "MFA session token from login (required)")
	verifyMFACmd.Flags().String("code", "", "TOTP or email OTP code (required)")

	_ = verifyMFACmd.MarkFlagRequired("mfa-session-token")
	_ = verifyMFACmd.MarkFlagRequired("code")

	rootCmd.AddCommand(verifyMFACmd)
}
