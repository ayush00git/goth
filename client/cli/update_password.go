package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

var updatePasswordCmd = &cobra.Command{
	Use:   "update-password",
	Short: "Change the authenticated user's password (logs out all sessions)",
	RunE: func(cmd *cobra.Command, args []string) error {
		accessToken, _ := cmd.Flags().GetString("access-token")
		userID, _ := cmd.Flags().GetString("user-id")
		oldPassword, _ := cmd.Flags().GetString("old-password")
		newPassword, _ := cmd.Flags().GetString("new-password")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		md := metadata.Pairs("authorization", "Bearer "+accessToken)
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		resp, err := client.UpdatePassword(ctx, &goth.UpdatePasswordRequest{
			UserId:      userID,
			OldPassword: oldPassword,
			NewPassword: newPassword,
		})
		if err != nil {
			return err
		}

		fmt.Println("Password Updated")
		fmt.Printf("Success:      %t\n", resp.Success)
		return nil
	},
}

func init() {
	updatePasswordCmd.Flags().String("access-token", "", "Bearer access token (required)")
	updatePasswordCmd.Flags().String("user-id", "", "User id (optional, must match the token)")
	updatePasswordCmd.Flags().String("old-password", "", "Current password (required)")
	updatePasswordCmd.Flags().String("new-password", "", "New password (required)")

	_ = updatePasswordCmd.MarkFlagRequired("access-token")
	_ = updatePasswordCmd.MarkFlagRequired("old-password")
	_ = updatePasswordCmd.MarkFlagRequired("new-password")

	rootCmd.AddCommand(updatePasswordCmd)
}
