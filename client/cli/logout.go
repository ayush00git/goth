package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Invalidate a refresh token (log out a device)",
	RunE: func(cmd *cobra.Command, args []string) error {
		refreshToken, _ := cmd.Flags().GetString("refresh-token")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		resp, err := client.Logout(context.Background(), &goth.LogoutRequest{
			RefreshToken: refreshToken,
		})
		if err != nil {
			return err
		}

		fmt.Println("Logout Successful")
		fmt.Printf("Success: %t\n", resp.Success)
		return nil
	},
}

func init() {
	logoutCmd.Flags().String("refresh-token", "", "Refresh token to invalidate (required)")
	_ = logoutCmd.MarkFlagRequired("refresh-token")

	rootCmd.AddCommand(logoutCmd)
}
