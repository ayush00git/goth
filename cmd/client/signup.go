package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"github.com/ayush00git/goth/grpc/goth"
)

func TestSignup(c goth.GothServiceClient) {
	// create a random number so that email never conflicts
	randomInt := rand.IntN(9999)

	resp, err := c.Signup(context.Background(), &goth.SignupRequest{
		Email: fmt.Sprintf("tester%d@gmail.com", randomInt),
		Password: "password123",
		FullName: "tester",
	})
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Signup Successfull\n");
	fmt.Printf("UserID: %s\n", resp.UserId);
	fmt.Printf("Email: %s\n", resp.Email);
	fmt.Printf("User Verified: %v\n", resp.EmailVerified);
}
