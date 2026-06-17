package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ayush00git/goth/grpc/goth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// connect to the gRPC server.
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodecV2(goth.CodecV2{}),
		),
	)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer conn.Close()

	// create the generated client
	client := goth.NewGothServiceClient(conn)
	resp, err := client.Signup(context.Background(), &goth.SignupRequest{
		Email: "test1@gmail.com",
		FullName: "tester",
		Password: "tester123",
	})
	if err != nil {
		log.Fatalf("Signup failed: %v", err)
	}

	fmt.Printf("User ID: %s\n", resp.UserId)
	fmt.Printf("Email: %s\n", resp.Email)
	fmt.Printf("Verified: %v\n", resp.EmailVerified)
}
