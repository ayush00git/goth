package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
)

var resetPasswordCmd = &cobra.Command{
	Use:   "reset-password",
	Short: "Set a new password with the session token from verify-otp --purpose 1 (logs out all sessions)",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionToken, _ := cmd.Flags().GetString("session-token")
		newPassword, _ := cmd.Flags().GetString("new-password")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		resp, err := client.ResetPassword(context.Background(), &goth.ResetPasswordRequest{
			SessionToken: sessionToken,
			NewPassword:  newPassword,
		})
		if err != nil {
			return err
		}

		fmt.Println("Password Reset")
		fmt.Printf("Success:      %t\n", resp.Success)
		return nil
	},
}

func init() {
	resetPasswordCmd.Flags().String("session-token", "", "Reset session token from verify-otp --purpose 1 (required)")
	resetPasswordCmd.Flags().String("new-password", "", "New password (required)")

	_ = resetPasswordCmd.MarkFlagRequired("session-token")
	_ = resetPasswordCmd.MarkFlagRequired("new-password")

	rootCmd.AddCommand(resetPasswordCmd)
}
