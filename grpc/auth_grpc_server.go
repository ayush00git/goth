package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ayush00git/goth/helpers"
	"github.com/ayush00git/goth/models"
	pb "github.com/ayush00git/goth/proto"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCServer struct {
	pb.UnimplementedAuthServiceServer
	Collection *mongo.Collection
}

func (s *AuthGRPCServer) Signup (ctx context.Context, req *pb.SignupRequest) (*pb.SignupResponse, error) {

	// instead of ShouldBondJSON
	user := models.User{
		ID: 			primitive.NewObjectID(),
		UserName: 		req.UserName,
		Email: 			req.Email,
		Role:  			"user",
		Password: 		req.Password,
		IsVerified: 	false,
		CreatedAt: 		time.Now(),
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, status.Error(codes.Internal, "error hashing password")
	}
	user.Password = string(hashPass)

	tokenString, err := helpers.GenerateToken(user.ID.Hex(), req.Email, req.UserName, user.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, "unable to generate a jwt token at the moment")
	}

	verificationURL := fmt.Sprintf("http://localhost:8080/api/auth/verify?token=%s", tokenString)
	go func() {
		helpers.SendEmail(req.Email, "email verification", verificationURL)
	} ()

	_, err = s.Collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			if strings.Contains(err.Error(), "email") {
				return nil, status.Error(codes.AlreadyExists, "email already in use")
			}
			if strings.Contains(err.Error(), "userName") {
				return nil, status.Error(codes.AlreadyExists, "username is already taken")
			}
		}
	}

	return &pb.SignupResponse{
		Message: "Check your inbox to verify the email",
		User: &pb.User{
			Id: user.ID.Hex(),
			UserName: user.UserName,
			Email: user.Email,
			Role: user.Role,
		},
	}, nil
}

func (s *AuthGRPCServer) Login (ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var foundUser models.User

	filter := bson.M{"email": req.Email}
	err := s.Collection.FindOne(ctx, filter).Decode(&foundUser)
	if err != nil {
		return nil, status.Error(codes.NotFound, "email not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(req.Password), []byte(foundUser.Password));
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "incorrect password")
	}

	tokenString, err := helpers.GenerateToken(foundUser.ID.Hex(), foundUser.Email, foundUser.UserName, foundUser.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, "error generating a jwt token")
	}

	return &pb.LoginResponse{
		Message: "Logged in successfully!",
		Token: tokenString,
	}, nil
}
