package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	).Scan(&user.id, &user.email, &user.password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found")
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
