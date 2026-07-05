package main

import (
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
	// TestSignup(client)
	// TestLogin(client)
	// TestLogout(client)
	// TestRefreshToken(client)

	// metadata to test the gRPC server without an actual gRPC client
	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYjc5MjU2MWEtNDAwZC00ZWQ3LWI0YWQtNWFlMTc4MDFlYzEyIiwiZW1haWwiOiJ0ZXN0ZXIyMTU3QHRlc3R0LmNvbSIsInN1YiI6ImI3OTI1NjFhLTQwMGQtNGVkNy1iNGFkLTVhZTE3ODAxZWMxMiIsImV4cCI6MTc4MzI1NDg1NiwiaWF0IjoxNzgzMjUzOTU2fQ.25m3YA4bLgVcHhh1yKnlSJg2IimE0bxc7EWudLTzYNM"

	TestEnableMFA(client, accessToken)
}
