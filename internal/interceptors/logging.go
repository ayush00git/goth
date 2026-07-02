package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

// LoggingInterceptor writes a log entry for every unary RPC request and response.
// It logs the method name, request duration, and whether the call returned an error.
func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	log.Printf("gRPC request started: method=%s", info.FullMethod)

	resp, err := handler(ctx, req)
	duration := time.Since(start)

	if err != nil {
		log.Printf("gRPC request finished: method=%s duration=%s error=%v", info.FullMethod, duration, err)
	} else {
		log.Printf("gRPC request finished: method=%s duration=%s response=%T", info.FullMethod, duration, resp)
	}

	return resp, err
}
