package wrtc

import (
	"context"
	"fmt"

	"github.com/pion/webrtc/v4"
)

type WRTCOptions struct {
	ctx              context.Context
	signaler         signaler
	connectionConfig webrtc.Configuration
	OnMessage        func(msg webrtc.DataChannelMessage)
}

var DefaultWRTCOptions = WRTCOptions{
	ctx:              context.Background(),
	signaler:         NewStdIoSignaler(),
	connectionConfig: webrtc.Configuration{},
}

type WRTCOption func(*WRTCOptions)

func WithWebrtcConfigWRTCOption(connectionConfig webrtc.Configuration) WRTCOption {
	return func(options *WRTCOptions) {
		options.connectionConfig = connectionConfig
	}
}

func WithSignalerWRTCOption(signaler signaler) WRTCOption {
	return func(options *WRTCOptions) {
		options.signaler = signaler
	}
}

func WithContextWRTCOption(ctx context.Context) WRTCOption {
	return func(options *WRTCOptions) {
		options.ctx = ctx
	}
}

func OnMessageWRTCOption(onMessage func(msg webrtc.DataChannelMessage)) WRTCOption {
	return func(options *WRTCOptions) {
		options.OnMessage = onMessage
	}
}

type WRTCConnection struct {
	connection  *webrtc.PeerConnection
	datachannel *webrtc.DataChannel
	options     WRTCOptions
	ready       chan error
}

func newConnection(options ...WRTCOption) (*WRTCConnection, error) {
	wrtcOptions := DefaultWRTCOptions
	for _, option := range options {
		option(&wrtcOptions)
	}

	peerConnection, err := webrtc.NewPeerConnection(wrtcOptions.connectionConfig)
	if err != nil {
		return nil, fmt.Errorf("an error occured while creating peer connection: %w", err)
	}

	wrtcConnection := WRTCConnection{
		ready:      make(chan error),
		options:    wrtcOptions,
		connection: peerConnection,
	}
	return &wrtcConnection, nil
}

// Create a wrtc connection that will initiate the connection as an offerer
// and wait for a response from the answerer
func NewConnectionAsOfferer(options ...WRTCOption) (*WRTCConnection, error) {
	wrtcConnection, err := newConnection(options...)
	if err != nil {
		return nil, fmt.Errorf("an error occured when initializing connection: %w", err)
	}

	// Initialize the connection in a separate goroutine so we can cancel it
	go func() {
		datachannel, err := wrtcConnection.connection.CreateDataChannel("main", nil)
		if err != nil {
			wrtcConnection.ready <- fmt.Errorf("an error occured while creating datachannel: %w", err)
			return
		}
		wrtcConnection.datachannel = datachannel
		datachannel.OnOpen(func() {
			wrtcConnection.ready <- nil
		})
		datachannel.OnMessage(wrtcConnection.options.OnMessage)

		err = wrtcConnection.options.signaler.SendOfferAndReceiveAnswer(wrtcConnection.connection)
		if err != nil {
			wrtcConnection.ready <- fmt.Errorf("an error occured during signaling phase: %w", err)
			return
		}
	}()

	select {
	case err := <-wrtcConnection.ready:
		if err != nil {
			return nil, err
		}
	case <-wrtcConnection.options.ctx.Done():
		return nil, fmt.Errorf("connection as offerer canceled")
	}

	return wrtcConnection, nil
}

// Create a wrtc connection that will initiate the connection as an answerer
// and wait for an offer from the offerer
func NewConnectionAsAnswerer(options ...WRTCOption) (*WRTCConnection, error) {
	wrtcConnection, err := newConnection(options...)
	if err != nil {
		return nil, fmt.Errorf("an error occured when initializing connection: %w", err)
	}

	// Initialize the connection in a separate goroutine so we can cancel it
	go func() {
		wrtcConnection.connection.OnDataChannel(func(datachannel *webrtc.DataChannel) {
			wrtcConnection.datachannel = datachannel
			datachannel.OnOpen(func() { wrtcConnection.ready <- nil })
			datachannel.OnMessage(wrtcConnection.options.OnMessage)
		})

		err = wrtcConnection.options.signaler.ReceiveOfferAndSendAnswer(wrtcConnection.connection)
		if err != nil {
			wrtcConnection.ready <- fmt.Errorf("an error occured during signaling phase: %w", err)
			return
		}
	}()

	select {
	case err := <-wrtcConnection.ready:
		if err != nil {
			return nil, err
		}
	case <-wrtcConnection.options.ctx.Done():
		return nil, fmt.Errorf("connection as answerer canceled")
	}

	return wrtcConnection, nil
}
