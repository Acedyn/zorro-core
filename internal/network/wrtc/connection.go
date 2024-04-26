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
	onMessage func(msg webrtc.DataChannelMessage)
	onOpen    func()
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

func WithMainChannelWRTCOption(mainChannelOptions DatachannelOptions) WRTCOption {
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

var mainDatachannelName = "main"

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

// Create a datachannel, renegociate the connection and wait for the other peer to pick up
func (wrtcConnection *WRTCConnection) CreateDatachannel(ctx context.Context, label string, options DatachannelOptions) (*webrtc.DataChannel, error) {
	if label == mainDatachannelName {
		return nil, fmt.Errorf("label %s is reserved for the connection's main datachannel", label)
	}

	datachannel, err := wrtcConnection.connection.CreateDataChannel(mainDatachannelName, nil)
	if err != nil {
		return nil, fmt.Errorf("an error occured while creating datachannel: %w", err)
	}

	// OnOpen triggers when the datachannel connection is established
	datachannelReady := make(chan error)
	datachannel.OnOpen(func() {
		datachannelReady <- nil
		if options.onOpen != nil {
			options.onOpen()
		}
	})

	// Optional custom handler on message
	if options.onMessage != nil {
		datachannel.OnMessage(options.onMessage)
	}

	select {
	case <-datachannelReady:
		return datachannel, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("datachannel creation canceled")
	}
}

// Wait for the peer to create a datachannel, renegociate the connection and pick it up
func (wrtcConnection *WRTCConnection) ReceiveDatachannel(ctx context.Context, label string, options DatachannelOptions) (*webrtc.DataChannel, error) {
	if label == mainDatachannelName {
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
		if options.onOpen != nil {
			datachannel.OnOpen(options.onOpen)
		}
		// Optional custom handler on message
		if options.onMessage != nil {
			datachannel.OnMessage(options.onMessage)
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
		datachannel, err := wrtcConnection.connection.CreateDataChannel(mainDatachannelName, nil)
		if err != nil {
			wrtcConnection.ready <- fmt.Errorf("an error occured while creating datachannel: %w", err)
			return
		}
		wrtcConnection.datachannel = datachannel

		// OnOpen triggers when the datachannel connection is established
		datachannel.OnOpen(func() {
			wrtcConnection.ready <- nil
			if wrtcConnection.options.mainChannel.onOpen != nil {
				wrtcConnection.options.mainChannel.onOpen()
			}
		})

		// Optional custom handler on message
		if wrtcConnection.options.mainChannel.onMessage != nil {
			datachannel.OnMessage(wrtcConnection.options.mainChannel.onMessage)
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
		if label == mainDatachannelName {
			wrtcConnection.datachannel = datachannel

			// OnOpen triggers when the datachannel connection is established
			datachannel.OnOpen(func() {
				wrtcConnection.ready <- nil
				if wrtcConnection.options.mainChannel.onOpen != nil {
					wrtcConnection.options.mainChannel.onOpen()
				}
			})

			// Optional custom handler on message
			if wrtcConnection.options.mainChannel.onMessage != nil {
				datachannel.OnMessage(wrtcConnection.options.mainChannel.onMessage)
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
