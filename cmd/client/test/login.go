package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
)

func TestLogin(c goth.GothServiceClient) {
	resp, err := c.Login(context.Background(), &goth.LoginRequest{
		Email: "tester2157@testt.com",
		Password: "password123",
		DeviceFingerprint: "web-chrome",	// might use uuid.NewString() to test how flow works for multiple devices
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("login successfull\n")
	fmt.Printf("access_token: %s\n", resp.AccessToken)
	fmt.Printf("refresh_token: %s\n", resp.RefreshToken)
	// mfa-type
}
