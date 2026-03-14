package grpc

import (
	"context"
	"fmt"
	"strings"
	"io"
	"time"
	"log"

	"github.com/ayush00git/goth/helpers"
	"github.com/ayush00git/goth/models"
	pb "github.com/ayush00git/goth/proto"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.Password));
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

func (s *AuthGRPCServer) Logout (ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{
		Message: "Logged out successfully!",
	}, nil
}

func (s *AuthGRPCServer) VerifyEmail (ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	claims, err := helpers.VerifyToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	var updatedUser models.User

	filter := bson.M{"email": claims.Email}
	update := bson.M{"$set": bson.M{"is_verified": true}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err = s.Collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedUser)
	if err != nil {
		return nil, status.Error(codes.Internal, "error updating user")
	}

	return &pb.VerifyEmailResponse{
		Message: "Email verified successfully!",
	}, nil 
}

func (s *AuthGRPCServer) GetUsers(ctx context.Context, req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	var users []models.User
	cursor, err := s.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, status.Error(codes.Internal, "error fetching users")
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &users); err != nil {
		return nil, status.Error(codes.Internal, "error decoding users")
	}

	var pbUsers []*pb.User
	for _, u := range users {
		pbUsers = append(pbUsers, &pb.User{
			Id: 			u.ID.Hex(),
			UserName: 		u.UserName,
			Email: 			u.Email,
			Role: 			u.Email,
			IsVerified: 	u.IsVerified,
		})
	}

	return &pb.GetUsersResponse{
		Message: "Users fetched successfully!",
		Users: pbUsers,
	}, nil
}

//server streaming
func (s *AuthGRPCServer) WatchSessions(req *pb.WatchSessionsRequest, stream pb.AuthService_WatchSessionsServer) error {
	// Notice: no context.Context param — it's on stream.Context()
	// Notice: no return value other than error — you send via stream.Send()

	log.Printf("Client subscribed to sessions for user: %s", req.UserId)

	// Simulate session events being pushed over time.
	// In production: tail a MongoDB change stream, read from a Redis pub/sub, etc.
	events := []pb.SessionEvent{
		{EventType: "login",   SessionId: "sess_001", IpAddress: "192.168.1.1", OccurredAt: time.Now().Format(time.RFC3339)},
		{EventType: "login",   SessionId: "sess_002", IpAddress: "10.0.0.5",    OccurredAt: time.Now().Format(time.RFC3339)},
		{EventType: "expired", SessionId: "sess_001", IpAddress: "192.168.1.1", OccurredAt: time.Now().Format(time.RFC3339)},
		{EventType: "logout",  SessionId: "sess_002", IpAddress: "10.0.0.5",    OccurredAt: time.Now().Format(time.RFC3339)},
	}

	for _, event := range events {
		// Check if client disconnected before sending each event
		select {
		case <-stream.Context().Done():
			log.Println("Client disconnected from WatchSessions")
			return nil
		default:
		}

		if err := stream.Send(&event); err != nil {
			return status.Errorf(codes.Internal, "error sending event: %v", err)
		}

		time.Sleep(1 * time.Second) // simulate events arriving over time
	}

	return nil // closing this = server says "stream is done"
}

// client streaming
// ```
// grpcurl -plaintext -d '
// {"token": "tok_aaa"}
// {"token": "tok_bbb"}
// {"token": "tok_ccc"}
// ' localhost:50051 auth.AuthService/BulkRevokeTokens
// ```
func (s *AuthGRPCServer) BulkRevokeTokens(stream pb.AuthService_BulkRevokeTokensServer) error {
	// Notice: no request param — you receive via stream.Recv()
	// Notice: you reply ONCE at the end via stream.SendAndClose()

	revokedCount := int32(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client closed their side — send the single response back
			return stream.SendAndClose(&pb.RevokeTokenResponse{
				RevokedCount: revokedCount,
			})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "error receiving token: %v", err)
		}

		// In production: add token to a blacklist in Redis, mark in DB, etc.
		log.Printf("Revoking token: %s", req.Token)
		revokedCount++
	}
}

// Bidirectional-streaming
// ```
// grpcurl -plaintext -d '
// {"action": "who_logged_in", "payload": "last_24h"}
// {"action": "list_sessions",  "payload": "user_abc"}
// {"action": "failed_logins",  "payload": "user_abc"}
// ' localhost:50051 auth.AuthService/AuditStream
// 
// ```

func (s *AuthGRPCServer) AuditStream(stream pb.AuthService_AuditStreamServer) error {
	// Both sides can send/receive concurrently.
	// Pattern: read client messages in a loop, respond to each one.

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil // client closed their side, we're done
		}
		if err != nil {
			return status.Errorf(codes.Internal, "recv error: %v", err)
		}

		log.Printf("Audit action received: %s | payload: %s", req.Action, req.Payload)

		// Build a response for each incoming message
		description := fmt.Sprintf("Processed action '%s' with payload '%s'", req.Action, req.Payload)

		event := &pb.AuditEvent{
			Description: description,
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		if err := stream.Send(event); err != nil {
			return status.Errorf(codes.Internal, "send error: %v", err)
		}
	}
}
