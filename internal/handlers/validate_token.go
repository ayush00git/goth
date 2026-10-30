package handlers

import (
	"context"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/ayush00git/goth/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidateToken checks an access token sent in the request body and returns
// who it belongs to. Other services call it to authenticate their own requests,
// so an invalid token is a normal answer (valid = false), not an rpc error.
func (g *GothServer) ValidateToken(ctx context.Context, req *goth.ValidateTokenRequest) (*goth.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	claims, err := services.ValidateAccessToken(req.AccessToken)
	if err != nil || claims == nil {
		return &goth.ValidateTokenResponse{Valid: false}, nil
	}

	return &goth.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID,
		Email:  claims.Email,
		// access tokens are only meant to be issued after MFA is completed
		// (Login for users without MFA, VerifyMFA for users with it).
		MfaVerified: true,
	}, nil
}
