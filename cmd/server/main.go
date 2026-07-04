package main

import (
	"context"
	"log"
	"net"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/db"
	"github.com/ayush00git/goth/internal/handlers"
	"github.com/ayush00git/goth/internal/helpers"
	"github.com/ayush00git/goth/internal/interceptors"
	"github.com/ayush00git/goth/internal/redis"
	"github.com/joho/godotenv"

	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading environment variables: %v", err.Error())
	}

	DB_URI := helpers.GetEnvVar("DB_CONN_URI");
	// connect to database on server initialization
	// ** for now we're using unencrypted tcp connections **
	// ** while deployment install postgres with ssl configurations **
	// ** to be able to set sslmode=require to use a encrypted tcp **
	db, err := db.New(DB_URI)
	if err != nil {
		log.Fatalf("unable to connect to the database: %v", err)
	}
	defer db.Close()
	log.Println("connected to database")

	// connect to redis client
	rdb := redis.RedisClient()
	if err := rdb.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	// open a raw tcp connection on gRPC default port
	// this connection just send/receive bytes, encoding
	// and framing is handled by HTTP2.0 and gRPC-go which
	// sits on its top
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed listening at port: %v", err)
	}

	// set grpc.ServerOptions to use fory's custom CodecV2
	// to use fory as the serialization layer instead of the
	// default protobuf
	s := grpc.NewServer(
		grpc.ForceServerCodecV2(goth.CodecV2{}),
		grpc.ChainUnaryInterceptor(
			interceptors.RecoveryInterceptor,
			interceptors.LoggingInterceptor,
			interceptors.AuthInterceptor,
		),
	)
	
	// register the gRPC server
	goth.RegisterGothServiceServer(
		s,
		handlers.NewServer(db.Pool, rdb),
	)

	log.Println("Goth server running on port :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}