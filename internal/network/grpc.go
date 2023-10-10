package network

import (
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
)

var (
	grpcServer *grpc.Server
	once       sync.Once
)

// Getter for the grpc server singleton
func GrpcServer() *grpc.Server {
	once.Do(func() {
		grpcServer = grpc.NewServer()
	})

	return grpcServer
}

// Start the grpc server
func ServeGrpc(host string, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	if err = GrpcServer().Serve(listener); err != nil {
		return fmt.Errorf("failed to server grpc: %w", err)
	}
	return nil
}
