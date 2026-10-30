package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
)

var validateTokenCmd = &cobra.Command{
	Use:   "validate-token",
	Short: "Check an access token and show who it belongs to",
	RunE: func(cmd *cobra.Command, args []string) error {
		accessToken, _ := cmd.Flags().GetString("access-token")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		// public rpc: the token goes in the request body, not the auth header.
		resp, err := client.ValidateToken(context.Background(), &goth.ValidateTokenRequest{
			AccessToken: accessToken,
		})
		if err != nil {
			return err
		}

		fmt.Println("Token Validated")
		fmt.Printf("Valid:        %t\n", resp.Valid)
		if resp.Valid {
			fmt.Printf("User ID:      %s\n", resp.UserId)
			fmt.Printf("Email:        %s\n", resp.Email)
			fmt.Printf("MFA Verified: %t\n", resp.MfaVerified)
		}
		return nil
	},
}

func init() {
	validateTokenCmd.Flags().String("access-token", "", "Access token to validate (required)")

	_ = validateTokenCmd.MarkFlagRequired("access-token")

	rootCmd.AddCommand(validateTokenCmd)
}
