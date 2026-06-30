package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/services"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoginReqBody
type RequestedUser struct {
	id			string
	email		string
	password	string
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
	).Scan(&user.id, &user.email, &user.password)
	if err == pgx.ErrNoRows {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "database error")
	}

	// match the password
	err = bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "credentials do not match.")
	}

	// generate an access token on login
	accessToken, err := services.GenerateAccessToken(user.id, user.email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating an access token")
	}

	// generate a refresh token
	refreshToken, jti, err := services.GenerateRefreshToken(user.id, user.email, "go-client")
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generating a refresh token")
	}

	// save jti to the db table
	_, err = g.db.Exec(
		ctx,
		`INSERT INTO REFRESH_TOKENS (user_id, jti, device_fingerprint)
		VALUES ($1, $2, $3)`,
		user.id,
		jti,
		"go-client",	// for testing
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to store refresh token to database")
	}

	// just placeholders for now
	return &goth.LoginResponse{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
		MfaRequired: true,
		MfaSessionToken: "need-to-be-set",
		MfaType: 0,
	}, nil
}
