package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

var enableMFACmd = &cobra.Command{
	Use:   "enable-mfa",
	Short: "Enable MFA for the authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		accessToken, _ := cmd.Flags().GetString("access-token")
		mfaType, _ := cmd.Flags().GetInt32("mfa-type")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		md := metadata.Pairs("authorization", "Bearer "+accessToken)
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		resp, err := client.EnableMFA(ctx, &goth.EnableMFARequest{
			MfaType: goth.MFAType(mfaType),
		})
		if err != nil {
			return err
		}

		fmt.Println("MFA Enabled")
		fmt.Printf("Success:      %t\n", resp.Success)
		fmt.Printf("TOTP Secret:  %s\n", resp.TotpSecret)
		fmt.Printf("TOTP QR URL:  %s\n", resp.TotpQrUrl)
		fmt.Printf("OTP Expires:  %d\n", resp.OtpExpiresIn)
		return nil
	},
}

func init() {
	enableMFACmd.Flags().String("access-token", "", "Bearer access token (required)")
	enableMFACmd.Flags().Int32("mfa-type", 1, "MFA type (0=None, 1=TOTP, 2=EmailOTP)")

	_ = enableMFACmd.MarkFlagRequired("access-token")

	rootCmd.AddCommand(enableMFACmd)
}
