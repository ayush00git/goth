package handlers

import (
	"github.com/ayush00git/goth/grpc/goth"
	"github.com/jackc/pgx/v5/pgxpool"
)

// embeds the generated grpc server implementation for forward compatibility.
// shared psql connection pools to be used by rpc handlers.
type GothServer struct {
	goth.UnimplementedGothServiceServer
	db *pgxpool.Pool
}

// NewServer creates a new gRPC sever with its required dependencies.
func NewServer(db *pgxpool.Pool) *GothServer {
	return &GothServer{
		db: db,
	}
}
