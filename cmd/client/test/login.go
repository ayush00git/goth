package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in as an existing user",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		device, _ := cmd.Flags().GetString("device")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		resp, err := client.Login(context.Background(), &goth.LoginRequest{
			Email:             email,
			Password:          password,
			DeviceFingerprint: device,
		})
		if err != nil {
			return err
		}

		fmt.Println("Login Successful")
		fmt.Printf("Access Token:  %s\n", resp.AccessToken)
		fmt.Printf("Refresh Token: %s\n", resp.RefreshToken)
		return nil
	},
}

func init() {
	loginCmd.Flags().String("email", "", "User email (required)")
	loginCmd.Flags().String("password", "", "Password (required)")
	loginCmd.Flags().String("device", "go-client", "Device fingerprint")

	_ = loginCmd.MarkFlagRequired("email")
	_ = loginCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(loginCmd)
}
