package grpc

import (
	"context"
	"net"
	"os"
	"os/signal"

	v1 "github.com/maslow123/go-grpc/pkg/api/v1"
	"github.com/maslow123/go-grpc/pkg/logger"
	"github.com/maslow123/go-grpc/pkg/protocol/grpc/middleware"
	"google.golang.org/grpc"
)

// RunServer runs gRPC service to publish Todo Service
func RunServer(ctx context.Context, v1API v1.TodoServiceServer, port string) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// gRPC server startup options
	opts := []grpc.ServerOption{}

	// add middleware
	opts = middleware.AddLogging(logger.Log, opts)

	// register service
	server := grpc.NewServer(opts...)
	v1.RegisterTodoServiceServer(server, v1API)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			// sig is a ^c, handle it
			logger.Log.Warn("Shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	// start gRPC server
	logger.Log.Info("Starting gRPC server...")
	return server.Serve(listen)
}
