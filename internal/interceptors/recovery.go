package interceptors

import (
	"context"
	"log"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor is a unary gRPC interceptor that protects the server
// from panic crashes by recovering and returning an internal error.
func RecoveryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("panic recovered in method %s: %v\n%s", info.FullMethod, rec, string(debug.Stack()))
			err = status.Error(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}
