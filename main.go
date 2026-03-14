package main

import (
	"log"
	"net"

	"github.com/ayush00git/goth/db"
	grpcserver "github.com/ayush00git/goth/grpc"
	"github.com/ayush00git/goth/handlers"
	pb "github.com/ayush00git/goth/proto"
	"github.com/ayush00git/goth/routes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc"

	"github.com/gin-gonic/gin"
)

func main() {

	AuthCollection := db.ConnectMongoDB("users")

	// Defining the handlers and routes
	db.CreateIndex(AuthCollection)

	authHandler := &handlers.AuthHandler{
		Collection: AuthCollection,
	}

	r := gin.Default()
	routes.AuthRoute(r, authHandler)

	go func() {
		log.Println("Server running on port 8080")
		log.Fatal(r.Run(":8080"))
	} ()

	// grpc server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to listed, err: %v", err)
	}

	grpcSrv := grpc.NewServer()

	pb.RegisterAuthServiceServer(grpcSrv, &grpcserver.AuthGRPCServer{
		Collection: AuthCollection,
	})

	reflection.Register(grpcSrv)
	log.Println("gRPC server running on :50051")
	log.Fatal(grpcSrv.Serve(lis))
}
