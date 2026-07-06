package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
)

var refreshTokenCmd = &cobra.Command{
	Use:   "refresh-token",
	Short: "Exchange a refresh token for a new access/refresh token pair",
	RunE: func(cmd *cobra.Command, args []string) error {
		refreshToken, _ := cmd.Flags().GetString("refresh-token")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		resp, err := client.RefreshToken(context.Background(), &goth.RefreshTokenRequest{
			RefreshToken: refreshToken,
		})
		if err != nil {
			return err
		}

		fmt.Println("Token Refreshed")
		fmt.Printf("Access Token:  %s\n", resp.AccessToken)
		fmt.Printf("Refresh Token: %s\n", resp.RefreshToken)
		return nil
	},
}

func init() {
	refreshTokenCmd.Flags().String("refresh-token", "", "Refresh token to exchange (required)")
	_ = refreshTokenCmd.MarkFlagRequired("refresh-token")

	rootCmd.AddCommand(refreshTokenCmd)
}
