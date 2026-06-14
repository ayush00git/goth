package main

import (
	"log"
	"net"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/handlers"
	"github.com/ayush00git/goth/internal/db"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	// connect to database on server initialization
	db, err := db.New("postgres://ayush00git:ayush00git@localhost:5432/goth?sslmode=disable")
	if err != nil {
		log.Fatalf("unable to connect to the database: %v", err)
	}
	defer db.Close()
	log.Println("connected to database")

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