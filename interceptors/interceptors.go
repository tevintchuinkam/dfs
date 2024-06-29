package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor adds a delay to all unary RPCs
func DelayInterceptor(delay time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// log.Printf("Applying delay of %s to unary call", delay)
		time.Sleep(delay)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor adds a delay to all stream RPCs
func DelayStreamInterceptor(delay time.Duration) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// log.Printf("Applying delay of %s to stream call", delay)
		time.Sleep(delay)
		return handler(srv, ss)
	}
}
