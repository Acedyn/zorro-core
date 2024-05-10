package rpc

import (
	"context"
	"fmt"

	"github.com/Acedyn/zorro-core/internal/network/wrtc"

	"github.com/pion/webrtc/v4"
)

type GrpcOverWrtcOptions struct {
	ctx                context.Context
	label              string
	asOfferer          bool
	datachannelOptions wrtc.DatachannelOptions
}

type GrpcOverWrtcOption func(*GrpcOverWrtcOptions)

var DefaultGrpcOverWrtcOptions = GrpcOverWrtcOptions{
	ctx:                context.Background(),
	label:              "grpc-over-wrtc",
	asOfferer:          false,
	datachannelOptions: wrtc.DatachannelOptions{},
}

// Pass a context to allow canceling
func WithContext(ctx context.Context) GrpcOverWrtcOption {
	return func(options *GrpcOverWrtcOptions) {
		options.ctx = ctx
	}
}

// Pass a custom label for the datachannel
func WithLabel(label string) GrpcOverWrtcOption {
	return func(options *GrpcOverWrtcOptions) {
		options.label = label
	}
}

// Configure if the peer should wait for the other peer to create the datachannel
// or create it himself
func WithAsOfferer(asOfferer bool) GrpcOverWrtcOption {
	return func(options *GrpcOverWrtcOptions) {
		options.asOfferer = asOfferer
	}
}

// Select custom datachannel options
func WithDatachannelOptions(datachannelOptions *wrtc.DatachannelOptions) GrpcOverWrtcOption {
	return func(options *GrpcOverWrtcOptions) {
		options.datachannelOptions = *datachannelOptions
	}
}

// Create the datachannel used to setup the grpc connection with the peer
func SetupGrpcOverWrtc(wrtcConnection *wrtc.WRTCConnection, options ...GrpcOverWrtcOption) (*webrtc.DataChannel, error) {
	grpcOverWrtcOption := DefaultGrpcOverWrtcOptions
	for _, option := range options {
		option(&grpcOverWrtcOption)
	}

	// Inject the server logic in the onMessage hook
	datachannelOptions := wrtc.DatachannelOptions{
		OnOpen: func() {
			// Call the custom hook after injecting the logic
			if grpcOverWrtcOption.datachannelOptions.OnOpen != nil {
				grpcOverWrtcOption.datachannelOptions.OnOpen()
			}
		},
		OnMessage: func(msg webrtc.DataChannelMessage) {
			// Call the custom hook after injecting the logic
			if grpcOverWrtcOption.datachannelOptions.OnMessage != nil {
				grpcOverWrtcOption.datachannelOptions.OnMessage(msg)
			}
		},
	}

	outputChan := make(chan struct {
		err         error
		datachannel *webrtc.DataChannel
	})

	// Create the datachannel in a different goroutine so we can wait for the wrtcConnection
	// to be established and wait for any cancelation at the same time
	go func() {
		if grpcOverWrtcOption.asOfferer {
			datachannel, err := wrtcConnection.CreateDatachannel(
				grpcOverWrtcOption.ctx,
				grpcOverWrtcOption.label,
				grpcOverWrtcOption.datachannelOptions,
			)
			outputChan <- struct {
				err         error
				datachannel *webrtc.DataChannel
			}{
				err:         err,
				datachannel: datachannel,
			}
		} else {
			datachannel, err := wrtcConnection.ReceiveDatachannel(
				grpcOverWrtcOption.ctx,
				grpcOverWrtcOption.label,
				datachannelOptions,
			)
			outputChan <- struct {
				err         error
				datachannel *webrtc.DataChannel
			}{
				err:         err,
				datachannel: datachannel,
			}
		}
	}()

	// Wait for the channel to be ready and for potiential cancellation at the same time
	select {
	case output := <-outputChan:
		if output.err != nil {
			return nil, fmt.Errorf("an error occured while seting up GRPC server over WRTC: %w", output.err)
		}
		return output.datachannel, nil
	case <-grpcOverWrtcOption.ctx.Done():
		return nil, fmt.Errorf("grpc over wrtc server setup canceled")
	}
}
