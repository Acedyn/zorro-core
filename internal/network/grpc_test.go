package network

import (
	"fmt"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// Test Starting the grpc server
func TestServeGrpc(t *testing.T) {
	host := "127.0.0.1"
	port, err := getFreePort()
	if err != nil {
		t.Errorf("Could not get free port: %s", err.Error())
	}

	// Start the server in its own goroutine
	go func() {
		if err := ServeGrpc(host, port); err != nil {
			t.Errorf("An error occured while serving GRPC: %s", err.Error())
		}
	}()
	grpcServer, _ := GrpcServer()
	defer grpcServer.GracefulStop()

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf("Could not start GRPC client: %s", err.Error())
	}
	if err = conn.Close(); err != nil {
		t.Errorf("Could not stop GRPC client: %s", err.Error())
	}
}
