package rpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Acedyn/zorro-core/internal/network/wrtc"
	test_rpc "github.com/Acedyn/zorro-core/testdata/rpc"
)

// Implemented echo service
type mockedEchoService struct {
	test_rpc.UnimplementedEchoServiceServer
}

func (service mockedEchoService) PingPong(ctx context.Context, in *test_rpc.Ping) (*test_rpc.Pong, error) {
	return &test_rpc.Pong{}, nil
}

func (service mockedEchoService) PingPongInfinite(in *test_rpc.Ping, stream test_rpc.EchoService_PingPongInfiniteServer) error {
	return nil
}

func (service mockedEchoService) PingPongCount(stream test_rpc.EchoService_PingPongCountServer) error {
	return nil
}

func (service mockedEchoService) PingPongInfiniteCount(stream test_rpc.EchoService_PingPongInfiniteCountServer) error {
	return nil
}

// Mocking of the sdp signaling
func mockedOnSendSdp(db map[string]chan string, key string) func(sdp string) error {
	return func(sdp string) error {
		db[key] <- sdp
		return nil
	}
}

func mockedOnReceiveSdp(db map[string]chan string, key string) func() (string, error) {
	return func() (string, error) {
		sdp := <-db[key]
		return sdp, nil
	}
}

func NewMockedSignaler() *wrtc.GenericSignaler {
	sdpDatabase := map[string]chan string{
		"offer":  make(chan string),
		"answer": make(chan string),
	}

	return &wrtc.GenericSignaler{
		OnSendOffer:     mockedOnSendSdp(sdpDatabase, "offer"),
		OnReceiveOffer:  mockedOnReceiveSdp(sdpDatabase, "offer"),
		OnSendAnswer:    mockedOnSendSdp(sdpDatabase, "answer"),
		OnReceiveAnswer: mockedOnReceiveSdp(sdpDatabase, "answer"),
	}

}

// Test creating a server and send unary requests to it with a client
func TestServerUnaryRequest(t *testing.T) {
	mockedSignaler := NewMockedSignaler()

	// The connection will be canceled after two seconds
	ctx, cancelContext := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelContext()

	// Create a server in its own goroutine
	serverChan := make(chan error)
	go func() {
		serverConnection, err := wrtc.NewConnectionAsOfferer(
			wrtc.WithSignalerWRTCOption(mockedSignaler),
			wrtc.WithContextWRTCOption(ctx),
		)
		if err != nil {
			serverChan <- err
			return
		}

		server, err := NewServer(
			serverConnection,
			WithContext(ctx),
			WithAsOfferer(true),
		)
		if err != nil {
			serverChan <- err
			return
		}

		test_rpc.RegisterEchoServiceServer(server, mockedEchoService{})
		serverChan <- nil
	}()

	// Create a client in its own goroutine
	clientChan := make(chan error)
	go func() {
		clientConnection, err := wrtc.NewConnectionAsAnswerer(
			wrtc.WithSignalerWRTCOption(mockedSignaler),
			wrtc.WithContextWRTCOption(ctx),
		)
		if err != nil {
			clientChan <- err
			return
		}

		grpcClientConnection, err := NewClient(
			clientConnection,
			WithContext(ctx),
		)
		if err != nil {
			clientChan <- err
			return
		}

		client := test_rpc.NewEchoServiceClient(grpcClientConnection)
		pong, err := client.PingPong(ctx, &test_rpc.Ping{})
		if err != nil {
			if pong == nil {
				err = fmt.Errorf("Invalid pong response: received nil instead of Pong{}")
			}
		}
		clientChan <- err
	}()

	// Wait for both goroutine to complete and check for errors
	for i := 0; i < 2; i += 1 {
		select {
		case err := <-serverChan:
			if err != nil {
				t.Errorf("An error occured on the server side connection: %s", err)
				cancelContext()
				return
			}
		case err := <-clientChan:
			if err != nil {
				t.Errorf("An error occured on the client side connection: %s", err)
				cancelContext()
				return
			}
		}
	}
}
