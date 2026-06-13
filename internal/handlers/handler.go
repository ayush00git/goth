package handlers

import (
	"github.com/ayush00git/goth/grpc/goth"
)

type Server struct {
	goth.UnimplementedGothServiceServer
}

func NewServer() *Server {
	return &Server{}
}
