package network

import (
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type GrpcServerStatus struct {
	Port      int
	Host      string
	IsRunning bool
}

var (
	grpcServerStatus *GrpcServerStatus
	grpcServer       *grpc.Server
	once             sync.Once
)

// Getter for the grpc server singleton
func GrpcServer() (*grpc.Server, *GrpcServerStatus) {
	once.Do(func() {
		grpcServer = grpc.NewServer()
		grpcServerStatus = &GrpcServerStatus{}
	})

	return grpcServer, grpcServerStatus
}

// Start the grpc server
func ServeGrpc(host string, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	server, status := GrpcServer()
	status.Port = port
	status.Host = host
	status.IsRunning = true

	if err = server.Serve(listener); err != nil {
		return fmt.Errorf("failed to server grpc: %w", err)
	}

	status.IsRunning = false
	return nil
}
