package handlers

import (
	"context"
	"log"

	"github.com/ayush00git/goth/grpc/goth"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		`INSERT INTO USERS (email, password_hash, full_name, email_verified)
		VALUES ($1, $2, $3, FALSE)
		RETURNING id`,
		email,
		hashedPass,
		fullName,
	).Scan(&userID)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create the user")
	}

	// Call SendOTP rpc handler with a flag 2.
	_, err = g.SendOTP(ctx, &goth.SendOTPRequest{
		Email: email,
		Purpose: goth.OTPPurposeEmailVerify,
	})
	if err != nil {
    	log.Printf("signup: failed sending verification otp to %s: %v", email, err)
	}

	// return the SignupResponse struct
	return &goth.SignupResponse{
		UserId: 		userID,
		Email: 			email,
		EmailVerified:  false,
	}, nil
}
