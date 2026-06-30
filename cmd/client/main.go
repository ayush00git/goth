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
	TestLogout(client)
}
