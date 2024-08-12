package wrtc

import (
	"context"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v4"
)

type WRTCOptions struct {
	ctx              context.Context
	signaler         signaler
	connectionConfig webrtc.Configuration
	mainChannel      DatachannelOptions
}

type DatachannelOptions struct {
	OnMessage func(msg webrtc.DataChannelMessage)
	OnOpen    func()
}

var DefaultWRTCOptions = WRTCOptions{
	ctx:              context.Background(),
	signaler:         NewStdIoSignaler(),
	connectionConfig: webrtc.Configuration{},
}

type WRTCOption func(*WRTCOptions)

// Allows passing custum webrtc config like ICE servers
func WithWebrtcConfigWRTCOption(connectionConfig webrtc.Configuration) WRTCOption {
	return func(options *WRTCOptions) {
		options.connectionConfig = connectionConfig
	}
}

// Pass a signaler that will handle the sdp exchange
func WithSignalerWRTCOption(signaler signaler) WRTCOption {
	return func(options *WRTCOptions) {
		options.signaler = signaler
	}
}

// Pass a context to allow canceling
func WithContextWRTCOption(ctx context.Context) WRTCOption {
	return func(options *WRTCOptions) {
		options.ctx = ctx
	}
}

// Hook and options for the default datachannel of the connection
func WithDefaultChannelWRTCOption(mainChannelOptions DatachannelOptions) WRTCOption {
	return func(options *WRTCOptions) {
		options.mainChannel = mainChannelOptions
	}
}

type WRTCConnection struct {
	connection                  *webrtc.PeerConnection
	datachannel                 *webrtc.DataChannel
	options                     WRTCOptions
	ready                       chan error
	newDatachannelsQueries      map[string]chan *webrtc.DataChannel
	newDatachannelsQueriesMutex sync.Mutex
	pendingNewDatachannels      map[string]*webrtc.DataChannel
	pendingNewDatachannelsMutex sync.Mutex
}

var defaultDatachannelName = "default"

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
		ready:                       make(chan error),
		options:                     wrtcOptions,
		connection:                  peerConnection,
		newDatachannelsQueries:      map[string]chan *webrtc.DataChannel{},
		newDatachannelsQueriesMutex: sync.Mutex{},
		pendingNewDatachannels:      map[string]*webrtc.DataChannel{},
		pendingNewDatachannelsMutex: sync.Mutex{},
	}
	return &wrtcConnection, nil
}

// Callback used when the wrtc connection receive a new datachannel
func (wrtcConnection *WRTCConnection) HandleNewDatachannel(datachannel *webrtc.DataChannel) {
	label := datachannel.Label()
	wrtcConnection.newDatachannelsQueriesMutex.Lock()
	defer wrtcConnection.newDatachannelsQueriesMutex.Unlock()

	if datachannelQuery, ok := wrtcConnection.newDatachannelsQueries[label]; ok {
		// If the datachannel was awaited, fullfill the query by sending it
		datachannelQuery <- datachannel
	} else {
		// If the datachannel was not awaited store it so it can be picked up
		wrtcConnection.pendingNewDatachannelsMutex.Lock()
		wrtcConnection.pendingNewDatachannels[label] = datachannel
		wrtcConnection.pendingNewDatachannelsMutex.Unlock()
	}
}

// Create a datachannel and wait for the other peer to pick up
func (wrtcConnection *WRTCConnection) CreateDatachannel(ctx context.Context, label string, options DatachannelOptions) (*webrtc.DataChannel, error) {
	if label == defaultDatachannelName {
		return nil, fmt.Errorf("label %s is reserved for the connection's main datachannel", label)
	}

	datachannel, err := wrtcConnection.connection.CreateDataChannel(label, nil)
	if err != nil {
		return nil, fmt.Errorf("an error occured while creating datachannel: %w", err)
	}

	// OnOpen triggers when the datachannel connection is established
	datachannelReady := make(chan error)
	datachannel.OnOpen(func() {
		datachannelReady <- nil
		if options.OnOpen != nil {
			options.OnOpen()
		}
	})

	// Optional custom handler on message
	if options.OnMessage != nil {
		datachannel.OnMessage(options.OnMessage)
	}

	select {
	case <-datachannelReady:
		return datachannel, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("datachannel creation canceled")
	}
}

// Wait for the peer to create a datachannel and pick it up
func (wrtcConnection *WRTCConnection) ReceiveDatachannel(ctx context.Context, label string, options DatachannelOptions) (*webrtc.DataChannel, error) {
	if label == defaultDatachannelName {
		return nil, fmt.Errorf("label %s is reserved for the connection's main datachannel", label)
	}

	// Make sure a new datachannel is not added while we look for it
	wrtcConnection.newDatachannelsQueriesMutex.Lock()

	// Check it the datachannel was not already created
	wrtcConnection.pendingNewDatachannelsMutex.Lock()
	if datachannel, ok := wrtcConnection.pendingNewDatachannels[label]; ok {
		wrtcConnection.pendingNewDatachannelsMutex.Unlock()
		wrtcConnection.newDatachannelsQueriesMutex.Unlock()
		return datachannel, nil
	}
	wrtcConnection.pendingNewDatachannelsMutex.Unlock()

	// Otherwire register a datachannel query so we can get it right after
	datachannelChan := make(chan *webrtc.DataChannel)
	wrtcConnection.newDatachannelsQueries[label] = datachannelChan
	wrtcConnection.newDatachannelsQueriesMutex.Unlock()

	select {
	case datachannel := <-datachannelChan:
		// Optional custom handler on open
		if options.OnOpen != nil {
			datachannel.OnOpen(options.OnOpen)
		}
		// Optional custom handler on message
		if options.OnMessage != nil {
			datachannel.OnMessage(options.OnMessage)
		}
		return datachannel, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("datachannel reception canceled")
	}
}

// Create a wrtc connection that will initiate the connection as an offerer
// and wait for a response from the answerer
func NewConnectionAsOfferer(options ...WRTCOption) (*WRTCConnection, error) {
	wrtcConnection, err := newConnection(options...)
	if err != nil {
		return nil, fmt.Errorf("an error occured when initializing connection: %w", err)
	}

	// Handle new datachannel requests
	wrtcConnection.connection.OnDataChannel(wrtcConnection.HandleNewDatachannel)

	// Initialize the connection in a separate goroutine so we can cancel it
	go func() {
		datachannel, err := wrtcConnection.connection.CreateDataChannel(defaultDatachannelName, nil)
		if err != nil {
			wrtcConnection.ready <- fmt.Errorf("an error occured while creating datachannel: %w", err)
			return
		}
		wrtcConnection.datachannel = datachannel

		// OnOpen triggers when the datachannel connection is established
		datachannel.OnOpen(func() {
			wrtcConnection.ready <- nil
			if wrtcConnection.options.mainChannel.OnOpen != nil {
				wrtcConnection.options.mainChannel.OnOpen()
			}
		})

		// Optional custom handler on message
		if wrtcConnection.options.mainChannel.OnMessage != nil {
			datachannel.OnMessage(wrtcConnection.options.mainChannel.OnMessage)
		}

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

	// The connection is considerer as ready when the main channel is active
	wrtcConnection.connection.OnDataChannel(func(datachannel *webrtc.DataChannel) {
		label := datachannel.Label()
		// Different handlers for the main channel and the other ones
		if label == defaultDatachannelName {
			wrtcConnection.datachannel = datachannel

			// OnOpen triggers when the datachannel connection is established
			datachannel.OnOpen(func() {
				wrtcConnection.ready <- nil
				if wrtcConnection.options.mainChannel.OnOpen != nil {
					wrtcConnection.options.mainChannel.OnOpen()
				}
			})

			// Optional custom handler on message
			if wrtcConnection.options.mainChannel.OnMessage != nil {
				datachannel.OnMessage(wrtcConnection.options.mainChannel.OnMessage)
			}
		} else {
			// Other datachannels will be handled differently
			wrtcConnection.HandleNewDatachannel(datachannel)
		}
	})

	// Initialize the connection in a separate goroutine so we can cancel it
	go func() {
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
