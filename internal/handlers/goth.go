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


func (g *GothServer) Login(ctx context.Context, req *goth.LoginRequest) (*goth.LoginResponse, error) {
	// inputs
	email := req.Email
	password := req.Password

	// validate the inputs
	if email == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required fields.")
	}

	// email lookup into db.
	var user RequestedUser
	err := g.db.QueryRow(
		ctx,
		`SELECT id, email, password_hash
		FROM users
		WHERE $1 = email`, email,
	).Scan(&user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch user.")
	}

	// match the password
	err = bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "credentials do not match.")
	}

	// just placeholders for now
	return &goth.LoginResponse{
		AccessToken: "need-to-be-set",
		RefreshToken: "need-to-be-set",
		MfaRequired: true,
		MfaSessionToken: "need-to-be-set",
		MfaType: 0,
	}, nil
}
