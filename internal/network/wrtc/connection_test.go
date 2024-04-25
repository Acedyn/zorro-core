package wrtc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"
)

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

// Test the connection between an answerer and an offerer
// in two different goroutines
func TestConnectionAsOffererAndAnswerer(t *testing.T) {
	sdpDatabase := map[string]chan string{
		"offer":  make(chan string),
		"answer": make(chan string),
	}

	mockedSignaler := genericSignaler{
		onSendOffer:     mockedOnSendSdp(sdpDatabase, "offer"),
		onReceiveOffer:  mockedOnReceiveSdp(sdpDatabase, "offer"),
		onSendAnswer:    mockedOnSendSdp(sdpDatabase, "answer"),
		onReceiveAnswer: mockedOnReceiveSdp(sdpDatabase, "answer"),
	}

	// The connection will be canceled after two seconds
	ctx, cancelContext := context.WithTimeout(context.Background(), time.Second*4)
	defer cancelContext()

	offererMessage := "ping"
	answererMessage := "pong"

	// Create the offerer connection on its own goroutine
	offererChan := make(chan error)
	go func() {
		messageChan := make(chan string)
		connection, err := NewConnectionAsOfferer(
			WithSignalerWRTCOption(&mockedSignaler),
			WithContextWRTCOption(ctx),
			OnMessageWRTCOption(func(msg webrtc.DataChannelMessage) {
				messageChan <- string(msg.Data)
			}),
		)
		if err != nil {
			offererChan <- err
			return
		}

		err = connection.datachannel.SendText(offererMessage)
		if err != nil {
			offererChan <- err
			return
		}

		receivedMessage := <-messageChan
		if receivedMessage != answererMessage {
			offererChan <- fmt.Errorf(
				"wrong message received, received: %s, expected: %s",
				receivedMessage,
				answererMessage,
			)
		}
		offererChan <- nil
	}()

	// Create the offerer connection on its own goroutine
	answererChan := make(chan error)
	go func() {
		messageChan := make(chan string)

		connection, err := NewConnectionAsAnswerer(
			WithSignalerWRTCOption(&mockedSignaler),
			WithContextWRTCOption(ctx),
			OnMessageWRTCOption(func(msg webrtc.DataChannelMessage) {
				messageChan <- string(msg.Data)
			}),
		)
		if err != nil {
			answererChan <- err
			return
		}

		receivedMessage := <-messageChan
		if receivedMessage != offererMessage {
			offererChan <- fmt.Errorf(
				"wrong message received, received %s, expected: %s",
				receivedMessage,
				offererMessage,
			)
		}

		err = connection.datachannel.SendText(answererMessage)
		if err != nil {
			answererChan <- err
			return
		}
		answererChan <- nil
	}()

	// Wait for both goroutine to complete and check for errors
	for i := 0; i < 2; i += 1 {
		select {
		case err := <-offererChan:
			if err != nil {
				t.Errorf("An error occured during wrtc offerer connection: %s", err)
				cancelContext()
				return
			}
		case err := <-answererChan:
			if err != nil {
				t.Errorf("An error occured during wrtc answerer connection: %s", err)
				cancelContext()
				return
			}
		}
	}
}
