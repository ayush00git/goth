package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

var sendOTPCmd = &cobra.Command{
	Use:   "send-otp",
	Short: "Send an OTP email (MFA, password reset or email verification)",
	Long: `Send an OTP email for one of three purposes:

  0 = MFA             needs --access-token (logged in) or --mfa-session-token (during login).
                      The email is taken from the token; --email is ignored.
  1 = Password reset  needs --email, no token.
  2 = Email verify    needs --email, no token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		purpose, _ := cmd.Flags().GetInt32("purpose")
		accessToken, _ := cmd.Flags().GetString("access-token")
		mfaSessionToken, _ := cmd.Flags().GetString("mfa-session-token")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		// attach only the tokens that were given, so the no-token path can be tested too.
		var pairs []string
		if accessToken != "" {
			pairs = append(pairs, "authorization", "Bearer "+accessToken)
		}
		if mfaSessionToken != "" {
			pairs = append(pairs, "x-mfa-session-token", mfaSessionToken)
		}
		ctx := context.Background()
		if len(pairs) > 0 {
			ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(pairs...))
		}

		resp, err := client.SendOTP(ctx, &goth.SendOTPRequest{
			Email:   email,
			Purpose: goth.OTPPurpose(purpose),
		})
		if err != nil {
			return err
		}

		fmt.Println("OTP Sent")
		fmt.Printf("Success:      %t\n", resp.Success)
		fmt.Printf("Expires In:   %ds\n", resp.ExpiresInSeconds)
		return nil
	},
}

func init() {
	sendOTPCmd.Flags().Int32("purpose", 0, "OTP purpose (0=MFA, 1=PasswordReset, 2=EmailVerify)")
	sendOTPCmd.Flags().String("email", "", "Target email (purpose 1 and 2)")
	sendOTPCmd.Flags().String("access-token", "", "Bearer access token (purpose 0, logged in)")
	sendOTPCmd.Flags().String("mfa-session-token", "", "MFA session token from login (purpose 0, during login)")

	rootCmd.AddCommand(sendOTPCmd)
}
