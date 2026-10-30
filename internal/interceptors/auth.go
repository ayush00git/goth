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

// optionalAuth rpcs accept requests with or without an access token.
// If a token is sent it must be valid, and its claims are attached to the
// context. The handler decides what each case is allowed to do.
var optionalAuth = map[string]bool{
	"/goth.GothService/SendOTP": true,
}

func AuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	// bypass AuthInterceptor for these public rpcs
	switch info.FullMethod {
		case "/goth.GothService/Login",
			"/goth.GothService/Signup",
			"/goth.GothService/VerifyOTP",
			"/goth.GothService/VerifyMFA",	// validates the mfa_session_token itself
			"/goth.GothService/ValidateToken":	// the token to check is in the request body
			return handler(ctx, req)
	}

	// read incoming metadata and the authorization header
	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get("authorization")
	if len(values) == 0 {
		// no token: let optional-auth rpcs through without claims
		if optionalAuth[info.FullMethod] {
			return handler(ctx, req)
		}
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
