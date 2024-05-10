package wrtc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
)

// Simple signaling method using stdin and stdout, the user will have to
// use copy paste to exchange the offer and answers
type stdIOSignaler struct {
	GenericSignaler
	ctx    context.Context
	reader io.Reader
	writer io.Writer
}

// Send the sdp via the signaler's writer
func onSendSdpStdIO(stdIOSignaler *stdIOSignaler, initMessage string) func(sdp string) error {
	return func(sdp string) error {
		_, err := stdIOSignaler.writer.Write([]byte(fmt.Sprintf("%s\n%s\n", initMessage, sdp)))
		return err
	}
}

// Get the sdp via the signaler's reader
func onReceiveSdpStdIO(stdIOSignaler *stdIOSignaler, initMessage string) func() (string, error) {
	return func() (string, error) {
		// We read for the stdin in a separated coroutine so we can wait for
		// cancelation at the same time
		stdinChan := make(chan struct {
			string
			error
		})
		go func() {
			stdinReader := bufio.NewReader(stdIOSignaler.reader)
			stdIOSignaler.writer.Write([]byte(fmt.Sprintf("%s\n", initMessage)))

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
				return "", fmt.Errorf("an error occured while reading stdin: %w", stdinSdp.error)
			}

			return stdinSdp.string, nil
		case <-stdIOSignaler.ctx.Done():
			return "", fmt.Errorf("sdp read from stdin canceled")
		}
	}
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

	stdIOSignaler.GenericSignaler = GenericSignaler{
		OnSendOffer:     onSendSdpStdIO(stdIOSignaler, "SDP OFFER:"),
		OnReceiveOffer:  onReceiveSdpStdIO(stdIOSignaler, "SDP OFFER:"),
		OnSendAnswer:    onSendSdpStdIO(stdIOSignaler, "SDP ANSWER:"),
		OnReceiveAnswer: onReceiveSdpStdIO(stdIOSignaler, "SDP ANSWER:"),
	}

	return stdIOSignaler
}
