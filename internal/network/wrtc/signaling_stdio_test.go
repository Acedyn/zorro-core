package wrtc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"
)

var webrtcConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

// Test the stdout signaling from the offerer point of view
func TestStdioOffererSignaling(t *testing.T) {
	offererPeerConnection, err := webrtc.NewPeerConnection(webrtcConfig)
	if err != nil {
		t.Errorf("An error occured while creating peer connection: %s", err)
	}
	_, err = offererPeerConnection.CreateDataChannel("main", nil)
	if err != nil {
		t.Errorf("An error occured while creating datachannel: %s", err)
	}

	answererPeerConnection, err := webrtc.NewPeerConnection(webrtcConfig)
	if err != nil {
		t.Errorf("An error occured while creating peer connection: %s", err)
	}

	// The signaling will be canceled after two seconds
	ctx, cancelContext := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelContext()

	// Create the stdio signaling
	pipeReader, pipeWriter := io.Pipe()
	stdIOSignaler := NewStdIoSignaling(
		offererPeerConnection,
		WithContext(ctx),
		WithReader(pipeReader),
		WithWriter(pipeWriter),
	)

	// Start the signaling process in a separate goroutine so we can
	// communicate with it via the reader and writer
	signalingChan := make(chan error)
	go func() {
		signalingChan <- stdIOSignaler.SendOfferAndReceiveAnswer()
	}()

	// Communicate with the signaling process via the reader and the
	// writer in a different goroutine so we can wait for both finishes
	testChan := make(chan error)
	go func() {
		// Create the reader to read the signaling io
		bufferReader := bufio.NewReader(pipeReader)
		// The signaler will send a message to announce the sdp offer
		initOfferMessage, err := bufferReader.ReadString('\n')
		if err != nil {
			testChan <- err
			return
		}
		if strings.Trim(initOfferMessage, "\n") != "SDP OFFER:" {
			testChan <- fmt.Errorf("init message not received (received %s)", initOfferMessage)
			return
		}

		// Read and use the actual offer in the second stream
		rawOffer, err := bufferReader.ReadString('\n')
		if err != nil {
			testChan <- err
			return
		}
		offer, err := decodeSdp(rawOffer)
		if err != nil {
			testChan <- err
			return
		}

		err = answererPeerConnection.SetRemoteDescription(*offer)
		if err != nil {
			testChan <- err
			return
		}

		// Create and send the answer
		answer, err := answererPeerConnection.CreateAnswer(nil)
		if err != nil {
			testChan <- err
			return
		}
		err = answererPeerConnection.SetLocalDescription(answer)
		if err != nil {
			testChan <- err
			return
		}

		encodedAnswer, err := encodeSdp(answer)
		if err != nil {
			testChan <- err
			return
		}

		// Just like the offer, the signaler will send a message to annouce
		// that he is ready to accept the answer
		initAnswerMessage, err := bufferReader.ReadString('\n')
		if err != nil {
			testChan <- err
			return
		}
		if strings.Trim(initAnswerMessage, "\n") != "SDP ANSWER:" {
			testChan <- fmt.Errorf("init message not received (received %s)", initAnswerMessage)
			return
		}

		pipeWriter.Write([]byte(fmt.Sprintf("%s\n", encodedAnswer)))
		testChan <- nil
	}()

	// Wait for both goroutine to complete and check for errors
	for i := 0; i < 2; i += 1 {
		select {
		case err := <-signalingChan:
			if err != nil {
				t.Errorf("An error occured during sdp exchange: %s", err)
				return
			}
		case err := <-testChan:
			if err != nil {
				t.Errorf("An error occured during sdp exchange tests: %s", err)
				cancelContext()
				return
			}
		}

		if signalingChan == nil && testChan == nil {
			break
		}
	}

}

// Test the stdout signaling from the answerer point of view
func TestStdioAnswererSignaling(t *testing.T) {
	// Create the webrtc connection
	answererPeerConnection, err := webrtc.NewPeerConnection(webrtcConfig)
	if err != nil {
		t.Errorf("An error occured while creating peer connection: %s", err)
	}

	offererPeerConnection, err := webrtc.NewPeerConnection(webrtcConfig)
	if err != nil {
		t.Errorf("An error occured while creating peer connection: %s", err)
	}
	_, err = offererPeerConnection.CreateDataChannel("main", nil)
	if err != nil {
		t.Errorf("An error occured while creating datachannel: %s", err)
	}

	// The signaling will be canceled after two seconds
	ctx, cancelContext := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelContext()

	// Create the stdio signaling
	pipeReader, pipeWriter := io.Pipe()
	stdIOSignaler := NewStdIoSignaling(
		answererPeerConnection,
		WithContext(ctx),
		WithReader(pipeReader),
		WithWriter(pipeWriter),
	)

	// Start the signaling process in a separate goroutine so we can
	// communicate with it via the reader and writer
	signalingChan := make(chan error)
	go func() {
		signalingChan <- stdIOSignaler.ReceiveOfferAndSendAnswer()
	}()

	// Communicate with the signaling process via the reader and the
	// writer in a different goroutine so we can wait for both finishes
	testChan := make(chan error)
	go func() {
		// Create the offer so we can send it to the answerer
		offer, err := offererPeerConnection.CreateOffer(nil)
		if err != nil {
			testChan <- err
			return
		}
		err = offererPeerConnection.SetLocalDescription(offer)
		if err != nil {
			testChan <- err
			return
		}

		encodedOffer, err := encodeSdp(offer)
		if err != nil {
			testChan <- err
			return
		}

		// Create the reader to read the signaling io
		bufferReader := bufio.NewReader(pipeReader)

		// The signaler will send a message to annouce
		// that he is ready to accept the offer
		initOfferMessage, err := bufferReader.ReadString('\n')
		if err != nil {
			testChan <- err
			return
		}
		if strings.Trim(initOfferMessage, "\n") != "SDP OFFER:" {
			testChan <- fmt.Errorf("init message not received (received %s)", initOfferMessage)
			return
		}

		pipeWriter.Write([]byte(fmt.Sprintf("%s\n", encodedOffer)))

		// The signaler will send a message to announce the sdp answer
		initAnswerMessage, err := bufferReader.ReadString('\n')
		if err != nil {
			testChan <- err
			return
		}
		if strings.Trim(initAnswerMessage, "\n") != "SDP ANSWER:" {
			testChan <- fmt.Errorf("init message not received (received %s)", initAnswerMessage)
			return
		}

		// Read and use the actual offer in the second stream
		rawAnswer, err := bufferReader.ReadString('\n')
		if err != nil {
			testChan <- err
			return
		}
		answer, err := decodeSdp(rawAnswer)
		if err != nil {
			testChan <- err
			return
		}

		err = offererPeerConnection.SetRemoteDescription(*answer)
		if err != nil {
			testChan <- err
			return
		}
		testChan <- nil
	}()

	// Wait for both goroutine to complete and check for errors
	for i := 0; i < 2; i += 1 {
		select {
		case err := <-signalingChan:
			if err != nil {
				t.Errorf("An error occured during sdp exchange: %s", err)
				return
			}
		case err := <-testChan:
			if err != nil {
				t.Errorf("An error occured during sdp exchange tests: %s", err)
				cancelContext()
				return
			}
		}

		if signalingChan == nil && testChan == nil {
			break
		}
	}
}
