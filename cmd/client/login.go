package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
)

func TestLogin(c goth.GothServiceClient) {
	_, err := c.Login(context.Background(), &goth.LoginRequest{
		Email: "tester6412@gmail.com",
		Password: "password123",
		DeviceFingerprint: "web-chrome",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Login Successfull\n")
	// tokens
}
