//go:build js && wasm

package wasm

import (
	"context"
	"fmt"

	"github.com/Acedyn/zorro-core/internal/network/wrtc"
)

// Temporary function used until transferable datachannel is more widely
// supported in browsers
func workerTempOnSendSdp(db map[string]chan string, key string) func(sdp string) error {
	return func(sdp string) error {
		db[key] <- sdp
		return nil
	}
}

// Temporary function used until transferable datachannel is more widely
// supported in browsers
func workerTempOnReceiveSdp(db map[string]chan string, key string) func() (string, error) {
	return func() (string, error) {
		sdp := <-db[key]
		return sdp, nil
	}
}

// Temporary signaler used until transferable datachannel is more widely
// supported in browsers
func newWorkerTempSignaler() *wrtc.GenericSignaler {
	sdpDatabase := map[string]chan string{
		"offer":  make(chan string),
		"answer": make(chan string),
	}

	return &wrtc.GenericSignaler{
		OnSendOffer:     workerTempOnSendSdp(sdpDatabase, "offer"),
		OnReceiveOffer:  workerTempOnReceiveSdp(sdpDatabase, "offer"),
		OnSendAnswer:    workerTempOnSendSdp(sdpDatabase, "answer"),
		OnReceiveAnswer: workerTempOnReceiveSdp(sdpDatabase, "answer"),
	}

}

// Simple signaling method with a web worker using postMessage()
// to exchange SDPs
type workerSignaler struct {
	wrtc.GenericSignaler
	ctx      context.Context
	jsworker *JsWorker
}

// Get the sdp answer via the signaler
func onReceiveSdpAnswer(workerSignaler *workerSignaler) func() (string, error) {
	return func() (string, error) {
		// Wait for an sdp answer while waiting for cancel at the same time
		select {
		case sdp := <-workerSignaler.jsworker.receivedSdpAnswer:
			return sdp, nil
		case <-workerSignaler.ctx.Done():
			return "", fmt.Errorf("sdp answer from web worker canceled")
		}
	}
}

// Get the sdp offer via the signaler
func onReceiveSdpOffer(workerSignaler *workerSignaler) func() (string, error) {
	return func() (string, error) {
		// Wait for an sdp offer while waiting for cancel at the same time
		select {
		case sdp := <-workerSignaler.jsworker.receivedSdpOffer:
			return sdp, nil
		case <-workerSignaler.ctx.Done():
			return "", fmt.Errorf("sdp offer from web worker canceled")
		}
	}
}

// Option pattern for optional parameters in the workerSignaler builder
type workerSignalerOption func(*workerSignaler)

func WithContextWasmOption(ctx context.Context) workerSignalerOption {
	return func(workerSignaler *workerSignaler) {
		workerSignaler.ctx = ctx
	}
}

// Public function to create a new workerSignaler
func NewWorkerSignaler(jsWorker *JsWorker, options ...workerSignalerOption) *workerSignaler {
	workerSignaler := &workerSignaler{
		ctx:      context.Background(),
		jsworker: jsWorker,
	}

	for _, option := range options {
		option(workerSignaler)
	}

	workerSignaler.GenericSignaler = wrtc.GenericSignaler{
		OnSendOffer:     workerSignaler.jsworker.SendSdpOffer,
		OnReceiveOffer:  onReceiveSdpOffer(workerSignaler),
		OnSendAnswer:    workerSignaler.jsworker.SendSdpAnswer,
		OnReceiveAnswer: onReceiveSdpAnswer(workerSignaler),
	}

	return workerSignaler
}
