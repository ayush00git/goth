package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

var getUserCmd = &cobra.Command{
	Use:   "get-user",
	Short: "Show the authenticated user's profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		accessToken, _ := cmd.Flags().GetString("access-token")
		userID, _ := cmd.Flags().GetString("user-id")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		md := metadata.Pairs("authorization", "Bearer "+accessToken)
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		resp, err := client.GetUser(ctx, &goth.GetUserRequest{
			UserId: userID,
		})
		if err != nil {
			return err
		}

		fmt.Println("User")
		fmt.Printf("User ID:        %s\n", resp.UserId)
		fmt.Printf("Email:          %s\n", resp.Email)
		fmt.Printf("Full Name:      %s\n", resp.FullName)
		fmt.Printf("Email Verified: %t\n", resp.EmailVerified)
		fmt.Printf("MFA Type:       %d\n", resp.MfaType)
		fmt.Printf("Created At:     %s\n", time.Unix(resp.CreatedAt, 0).Format(time.RFC3339))
		return nil
	},
}

func init() {
	getUserCmd.Flags().String("access-token", "", "Bearer access token (required)")
	getUserCmd.Flags().String("user-id", "", "User id (optional, must match the token)")

	_ = getUserCmd.MarkFlagRequired("access-token")

	rootCmd.AddCommand(getUserCmd)
}
