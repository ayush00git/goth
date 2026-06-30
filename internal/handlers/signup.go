package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// LoginReqBody
type RequestedUser struct {
	id			string
	email		string
	fullName	string
	password	string
}

func (g *GothServer) Signup(ctx context.Context, req *goth.SignupRequest) (*goth.SignupResponse, error) {
	// inputs from request body
	email := req.Email
	password := req.Password
	fullName := req.FullName

	// validate the requests
	if email == "" || password == "" || fullName == "" {
		return nil, status.Error(codes.InvalidArgument, "email, full_name and password are required fields")
	}

	// hash the password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash the password")
	}

	// write to db
	var userID string
	err = g.db.QueryRow(
		ctx,
		`INSERT INTO USERS (email, password_hash, full_name)
		VALUES ($1, $2, $3)
		RETURNING id`,
		email,
		hashedPass,
		fullName,
	).Scan(&userID)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create the user")
	}

	// return the SignupResponse struct
	return &goth.SignupResponse{
		UserId: 		userID,
		Email: 			email,
		EmailVerified:  false,
	}, nil
}
