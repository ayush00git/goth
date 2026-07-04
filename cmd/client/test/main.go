package main

import (
	"log"

	"github.com/ayush00git/goth/grpc/goth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// metadata to test the gRPC server without an actual gRPC client
	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiY2E1MTg1ZDAtMWU0Ni00ZDdhLWFjMGItMmE4M2IyNTU5NGRlIiwiZW1haWwiOiJ0ZXN0ZXI2NDEyQGdtYWlsLmNvbSIsInN1YiI6ImNhNTE4NWQwLTFlNDYtNGQ3YS1hYzBiLTJhODNiMjU1OTRkZSIsImV4cCI6MTc4MzE3NjEwNiwiaWF0IjoxNzgzMTc1MjA2fQ.5jf1GxA6YIitMnlvI7hw-DzpcSCcMSA7hH94zvLjRio"

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
	// TestSignup(client)
	// TestLogin(client)
	// TestLogout(client)
	// TestRefreshToken(client)
	TestEnableMFA(client, accessToken)
}
