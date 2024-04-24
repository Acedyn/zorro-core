package wrtc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pion/webrtc/v4"
)

// Simple signaling method using stdin and stdout, the user will have to
// use copy paste to exchange the offer and answers
type stdIOSignaler struct {
	ctx    context.Context
	reader io.Reader
	writer io.Writer
}

// Option pattern for optional parameters in the stdIOSignaler builder
type stdIOSignalerOption func(*stdIOSignaler)

func WithContextStdIOOption(ctx context.Context) stdIOSignalerOption {
	return func(stdIOSignaler *stdIOSignaler) {
		stdIOSignaler.ctx = ctx
	}
}

func WithReaderStdIOOption(reader io.Reader) stdIOSignalerOption {
	return func(stdIOSignaler *stdIOSignaler) {
		stdIOSignaler.reader = reader
	}
}

func WithWriterStdIOOption(writer io.Writer) stdIOSignalerOption {
	return func(stdIOSignaler *stdIOSignaler) {
		stdIOSignaler.writer = writer
	}
}

// Public function to create a new stdIOSignaler
func NewStdIoSignaler(options ...stdIOSignalerOption) *stdIOSignaler {
	stdIOSignaler := &stdIOSignaler{
		ctx:    context.Background(),
		reader: os.Stdin,
		writer: os.Stdout,
	}

	for _, option := range options {
		option(stdIOSignaler)
	}

	return stdIOSignaler
}

// Read the std in and expect a base64 encoded SDP offer or answer
func (stdIOSignaler *stdIOSignaler) setRemoteSDPFromStdin(peerConnection *webrtc.PeerConnection, printMessage string) error {
	// We read for the stdin in a separated coroutine so we can wait for
	// cancelation at the same time
	stdinChan := make(chan struct {
		string
		error
	})
	go func() {
		stdinReader := bufio.NewReader(stdIOSignaler.reader)
		stdIOSignaler.writer.Write([]byte(fmt.Sprintf("%s:\n", printMessage)))

		stdinSdp, err := stdinReader.ReadString('\n')
		stdinChan <- struct {
			string
			error
		}{stdinSdp, err}
	}()

	// Wait for an std input while waiting for cancel at the same time
	select {
	case stdinSdp := <-stdinChan:
		if stdinSdp.error != nil {
			return fmt.Errorf("an error occured while reading stdin: %w", stdinSdp.error)
		}

		decodedSdp, err := decodeSdp(stdinSdp.string)
		if err != nil {
			return fmt.Errorf("an error occured while decoding stdin: %w", err)
		}

		err = peerConnection.SetRemoteDescription(*decodedSdp)
		if err != nil {
			return fmt.Errorf("could not set remote description %v: %w", decodedSdp, err)
		}
	case <-stdIOSignaler.ctx.Done():
		return fmt.Errorf("sdp read from stdin canceled")
	}

	return nil
}

func (stdIOSignaler *stdIOSignaler) SendOfferAndReceiveAnswer(peerConnection *webrtc.PeerConnection) error {
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("could not create offer for stdout signaling: %w", err)
	}

	// The offer is always a local description
	peerConnection.SetLocalDescription(offer)

	encodedSdp, err := encodeSdp(offer)
	if err != nil {
		return fmt.Errorf("could not encode offer for stdout signaling: %w", err)
	}

	// The exchange of sdp is done using stdout by simply printing it
	stdIOSignaler.writer.Write([]byte(fmt.Sprintf("SDP OFFER:\n%s\n", encodedSdp)))

	err = stdIOSignaler.setRemoteSDPFromStdin(peerConnection, "SDP ANSWER")
	if err != nil {
		return fmt.Errorf("an error occured when reading answer from stdin: %w", err)
	}

	return nil
}

func (stdIOSignaler *stdIOSignaler) ReceiveOfferAndSendAnswer(peerConnection *webrtc.PeerConnection) error {
	err := stdIOSignaler.setRemoteSDPFromStdin(peerConnection, "SDP OFFER")
	if err != nil {
		return fmt.Errorf("an error occured when reading offer from stdin: %w", err)
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("could not create answer for stdout signaling: %w", err)
	}

	// The offer is always a local description
	peerConnection.SetLocalDescription(answer)

	encodedSdp, err := encodeSdp(answer)
	if err != nil {
		return fmt.Errorf("could not encode answer for stdout signaling: %w", err)
	}

	// The exchange of sdp is done using stdout by simply printing it
	stdIOSignaler.writer.Write([]byte(fmt.Sprintf("SDP ANSWER:\n%s\n", encodedSdp)))
	return nil
}
