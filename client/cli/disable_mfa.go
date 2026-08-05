package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

var disableMFACmd = &cobra.Command{
	Use:   "disable-mfa",
	Short: "Disable MFA for the authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		accessToken, _ := cmd.Flags().GetString("access-token")
		code, _ := cmd.Flags().GetString("code")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		md := metadata.Pairs("authorization", "Bearer "+accessToken)
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		resp, err := client.DisableMFA(ctx, &goth.DisableMFARequest{
			Code: code,
		})
		if err != nil {
			return err
		}

		fmt.Println("MFA Disabled")
		fmt.Printf("Success:      %t\n", resp.Success)
		return nil
	},
}

func init() {
	disableMFACmd.Flags().String("access-token", "", "Bearer access token (required)")
	disableMFACmd.Flags().String("code", "", "TOTP or email OTP code (required)")

	_ = disableMFACmd.MarkFlagRequired("access-token")
	_ = disableMFACmd.MarkFlagRequired("code")

	rootCmd.AddCommand(disableMFACmd)
}
