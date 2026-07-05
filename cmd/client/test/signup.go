package main

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
)

var signupCmd = &cobra.Command{
	Use:   "signup",
	Short: "Register a new user",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		name, _ := cmd.Flags().GetString("name")
		password, _ := cmd.Flags().GetString("password")
		random, _ := cmd.Flags().GetBool("random-email")

		client, conn, err := grpcClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		if random {
			email = fmt.Sprintf("tester%d@testt.com", rand.IntN(9999))
		}

		resp, err := client.Signup(context.Background(), &goth.SignupRequest{
			Email:    email,
			Password: password,
			FullName: name,
		})
		if err != nil {
			return err
		}

		fmt.Println("Signup Successful")
		fmt.Printf("UserID:         %s\n", resp.UserId)
		fmt.Printf("Email:          %s\n", resp.Email)
		fmt.Printf("Email Verified: %v\n", resp.EmailVerified)
		return nil
	},
}

func init() {
	signupCmd.Flags().String("email", "", "User email")
	signupCmd.Flags().String("name", "", "Full name (required)")
	signupCmd.Flags().String("password", "", "Password (required)")
	signupCmd.Flags().Bool("random-email", false, "Generate a random email (ignores --email)")

	_ = signupCmd.MarkFlagRequired("name")
	_ = signupCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(signupCmd)
}
