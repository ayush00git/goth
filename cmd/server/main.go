package main

import (
	"log"
	"net"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/handlers"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// open a raw tcp connection on gRPC default port
	// this connection just send/receive bytes, encoding
	// and framing is handled by HTTP2.0 and gRPC-go which
	// sits on its top
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed listening at port: %v", err)
	}

	s := grpc.NewServer()
	// register the gRPC server
	goth.RegisterGothServiceServer(s, handlers.NewServer())
	reflection.Register(s)	// reflection package helps grpcurl recognize the goth service definitions

	log.Println("Goth server running on port :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}