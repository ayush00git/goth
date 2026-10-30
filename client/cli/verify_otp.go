package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
)

var verifyOTPCmd = &cobra.Command{
	Use:   "verify-otp",
	Short: "Verify an OTP sent by send-otp (MFA, password reset or email verification)",
	Long: `Verify an OTP for one of three purposes:

  0 = MFA             rejected by the server: use verify-mfa or disable-mfa instead.
  1 = Password reset  needs --email. Returns a session token for the reset step.
  2 = Email verify    needs --email. Marks the email as verified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		code, _ := cmd.Flags().GetString("code")
		purpose, _ := cmd.Flags().GetInt32("purpose")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		resp, err := client.VerifyOTP(context.Background(), &goth.VerifyOTPRequest{
			Email:   email,
			Code:    code,
			Purpose: goth.OTPPurpose(purpose),
		})
		if err != nil {
			return err
		}

		fmt.Println("OTP Verified")
		fmt.Printf("Success:       %t\n", resp.Success)
		if resp.SessionToken != "" {
			fmt.Printf("Session Token: %s\n", resp.SessionToken)
		}
		return nil
	},
}

func init() {
	verifyOTPCmd.Flags().Int32("purpose", 1, "OTP purpose (1=PasswordReset, 2=EmailVerify; 0=MFA is rejected)")
	verifyOTPCmd.Flags().String("code", "", "OTP code from the email (required)")
	verifyOTPCmd.Flags().String("email", "", "Email the code was sent to (purpose 1 and 2)")

	_ = verifyOTPCmd.MarkFlagRequired("code")

	rootCmd.AddCommand(verifyOTPCmd)
}
