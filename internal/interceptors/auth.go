package interceptors

import (
	"context"
	"strings"

	"github.com/ayush00git/goth/internal/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	// bypass AuthInterceptor for these public rpcs
	switch info.FullMethod {
		case "/goth.GothService/Login",
			"/goth.GothService/Signup",
			"/goth.GothService/VerifyOTP":
			return handler(ctx, req)
	}

	// read incoming metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// read authorization header
	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header missing")
	}

	authHeader := values[0]

	// must have a prefix "Bearer"
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// validate the access token
	claims, err := services.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid access token")
	}

	// store the claims in context
	ctx = context.WithValue(ctx, UserContextKeys{}, claims)

	return handler(ctx, req)
}
