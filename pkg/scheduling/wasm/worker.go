//go:build js && wasm

package wasm

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"syscall/js"

	"github.com/Acedyn/zorro-core/internal/filesystem"
	"github.com/Acedyn/zorro-core/internal/network/wrtc"
	"github.com/Acedyn/zorro-core/internal/utils"

	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
	"github.com/life4/genesis/slices"
	"google.golang.org/protobuf/encoding/protojson"
)

type JsWorker struct {
	workerApi         js.Value
	blobApi           js.Value
	jsonApi           js.Value
	urlApi            js.Value
	worker            js.Value
	receivedSdpOffer  chan string
	receivedSdpAnswer chan string
	processor         *scheduling_proto.Processor
	path              *scheduling_proto.ProcessorPath

	// The connection lives here until transferable datachannels is more widely
	// supported among browsers
	Connection     *wrtc.WRTCConnection
	PendingForward chan *scheduling_proto.IpcPacket_WrtcForward
}

// Start a new web worker that will run the given processor's script
func NewWorker(processorToStart *scheduling_proto.Processor) (*JsWorker, error) {
	// Initialize the worker struct
	jsWorker := JsWorker{
		processor:         processorToStart,
		receivedSdpOffer:  make(chan string),
		receivedSdpAnswer: make(chan string),
		PendingForward:    make(chan *scheduling_proto.IpcPacket_WrtcForward),
	}
	err := jsWorker.getJsApis()
	if err != nil {
		return nil, err
	}

	// Start a worker that will execute the processor's script
	errors := []error{}
	for _, jsWorker.path = range processorToStart.GetPaths() {
		// Try every configured path to start the processor
		processorFileSystem, err := filesystem.GetFileSystem(jsWorker.path.GetFileSystem())
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// Read the script file
		fileHandle, err := processorFileSystem.Open(jsWorker.path.GetStartProcessorTemplate())
		if err != nil {
			errors = append(errors, err)
			continue
		}
		defer fileHandle.Close()
		fileData, err := io.ReadAll(fileHandle)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// Wasm files are handled differently since they need to be instantiated
		isWasm := filepath.Ext(jsWorker.path.GetStartProcessorTemplate()) == ".wasm"

		// Store the file content in a blob
		processorRawContent := js.Global().Get("Uint8Array").New(len(fileData))
		js.CopyBytesToJS(processorRawContent, fileData)
		processorScriptFile := jsWorker.blobApi.New(
			processorRawContent,
			map[string]any{"type": "text/javascript"},
		)
		processorScriptUrl := jsWorker.urlApi.Call("createObjectURL", processorScriptFile)

		if isWasm {
			// Rebuild the blob with the right MIME types
			processorScriptFile = processorScriptFile.Call("slice", 0,
				processorScriptFile.Get("size"), "application/wasm",
			)
			processorScriptUrl = jsWorker.urlApi.Call("createObjectURL", processorScriptFile)

			// When the start processor template path is a wasm file, instantiate
			// the wasm module first
			processorScriptFile = jsWorker.blobApi.New(
				[]any{fmt.Sprintf("WebAssembly.instantiateStreaming(fetch('%s'))", processorScriptUrl)},
				map[string]any{"type": "text/javascript"},
			)
			js.Global().Get("WebAssembly").Call("instantiateStreaming", js.Global().Call("fetch", processorScriptUrl))
			processorScriptUrl = jsWorker.urlApi.Call("createObjectURL", processorScriptFile)
		}

		// Start the worker by giving it the dynamically created url
		jsWorker.worker = jsWorker.workerApi.New(processorScriptUrl)
		break
	}

	// If none of the proposed script path returns an error
	if jsWorker.worker.IsUndefined() {
		return nil, fmt.Errorf(
			"no valid processor could be started as a web worker: \n%w",
			fmt.Errorf(strings.Join(slices.Map(errors, func(e error) string { return e.Error() }), "\n")),
		)
	}

	// Add the handlers to the onmessage event
	jsWorker.worker.Set("onmessage", js.FuncOf(jsWorker.onWorkerMessage))
	return &jsWorker, nil
}

// Check and store handle to various js APIs
func (jsWorker *JsWorker) getJsApis() error {
	// Get and check the worker API
	workerApi := js.Global().Get("Worker")
	if workerApi.IsUndefined() {
		return fmt.Errorf("web worker api not supported")
	}
	jsWorker.workerApi = workerApi

	// Get and check the blob API
	blobApi := js.Global().Get("Blob")
	if blobApi.IsUndefined() {
		return fmt.Errorf("blob api not supported")
	}
	jsWorker.blobApi = blobApi

	// Get and check the blob API
	jsonApi := js.Global().Get("JSON")
	if jsonApi.IsUndefined() {
		return fmt.Errorf("json api not supported")
	}
	jsWorker.jsonApi = jsonApi

	// Get and check the url API
	urlApi := js.Global().Get("URL")
	if urlApi.IsUndefined() {
		return fmt.Errorf("url api not supported")
	}
	jsWorker.urlApi = urlApi

	return nil
}

// Log an error after appening infos about the worker
func (jsWorker *JsWorker) logError(msg string, args ...any) {
	args = append(args, jsWorker.processor.Name, jsWorker.processor.Id)
	utils.Logger().Error(fmt.Sprintf(msg+" on processor's worker %s:%s", args...))
}

// Message handler to transfer files and SDPs
func (jsWorker *JsWorker) onWorkerMessage(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		jsWorker.logError("Received message from web worker with %d argument instead of 1", len(args))
		return nil
	}

	// The wasm worker for a processor must follow a strict protocol defined here
	parsedData := scheduling_proto.IpcPacket{}
	err := protojson.Unmarshal([]byte(args[0].Get("data").String()), &parsedData)
	if err != nil {
		jsWorker.logError(
			"Received invalid message from web worker (%s)",
			jsWorker.jsonApi.Call("stringify", args[0].Get("data")).String(),
		)
		return nil
	}

	// Handle the incomming packet
	switch procedure := parsedData.PacketType.(type) {
	case *scheduling_proto.IpcPacket_InitPacket:
		// The client side is ignoring init packets
	case *scheduling_proto.IpcPacket_SdpExchange:
		switch sdp := procedure.SdpExchange.Sdp.(type) {
		case *scheduling_proto.SdpExchange_Offer:
			jsWorker.onSdpOffer(sdp)
		case *scheduling_proto.SdpExchange_Answer:
			jsWorker.onSdpAnswer(sdp)
		}
	case *scheduling_proto.IpcPacket_WrtcForward:
		jsWorker.onWrtcForward(procedure)
	}

	return nil
}

// Handler for sdp offer packets comming from the worker
func (jsWorker *JsWorker) onSdpOffer(offer *scheduling_proto.SdpExchange_Offer) error {
	// If an sdp was already pushed on the channel, replace it
	select {
	case <-jsWorker.receivedSdpOffer:
		jsWorker.receivedSdpOffer <- offer.Offer
	default:
		jsWorker.receivedSdpOffer <- offer.Offer
	}
	return nil
}

// Handler for sdp answer packets comming from the worker
func (jsWorker *JsWorker) onSdpAnswer(answer *scheduling_proto.SdpExchange_Answer) error {
	// If an sdp was already pushed on the channel, replace it
	select {
	case <-jsWorker.receivedSdpAnswer:
		jsWorker.receivedSdpAnswer <- answer.Answer
	default:
		jsWorker.receivedSdpAnswer <- answer.Answer
	}
	return nil
}

// Handler for wrtc forward packets comming from the worker
func (jsWorker *JsWorker) onWrtcForward(packet *scheduling_proto.IpcPacket_WrtcForward) error {
	// If the connection is not ready yet, we must make sure that the packet is processed
	// once the connection is ready. When settings the connection, HandlePendingForward
	// must be called
	jsWorker.PendingForward <- packet
	if jsWorker.Connection == nil {
		return nil
	}

	jsWorker.HandlePendingForward()
	return nil
}

// Handle the incomming forward packets that are pending. This function is separated
// from onWrtcForward to keep a queue of packets waiting until the wrtc connection
// is set
func (jsWorker *JsWorker) HandlePendingForward() error {
	if jsWorker.Connection == nil {
		return nil
	}

	select {
	case <-jsWorker.PendingForward:
	default:
	}

	return nil
}

// Send an SDP offer
func (jsWorker *JsWorker) SendSdpOffer(sdp string) error {
	ipcPacket := scheduling_proto.IpcPacket{
		PacketType: &scheduling_proto.IpcPacket_SdpExchange{
			SdpExchange: &scheduling_proto.SdpExchange{
				Sdp: &scheduling_proto.SdpExchange_Offer{
					Offer: sdp,
				},
			},
		},
	}

	bytes, err := protojson.Marshal(&ipcPacket)
	if err != nil {
		return fmt.Errorf("send offer IPC packet could not be marshaled: %w", err)
	}
	jsWorker.worker.Call("postMessage", string(bytes))
	return nil
}

// Send an SDP offer
func (jsWorker *JsWorker) SendSdpAnswer(sdp string) error {
	ipcPacket := scheduling_proto.IpcPacket{
		PacketType: &scheduling_proto.IpcPacket_SdpExchange{
			SdpExchange: &scheduling_proto.SdpExchange{
				Sdp: &scheduling_proto.SdpExchange_Answer{
					Answer: sdp,
				},
			},
		},
	}

	bytes, err := protojson.Marshal(&ipcPacket)
	if err != nil {
		return fmt.Errorf("send offer IPC packet could not be marshaled: %w", err)
	}
	jsWorker.worker.Call("postMessage", string(bytes))
	return nil
}

// Send a message to the worker indicating to start
func (jsWorker *JsWorker) Init() error {
	ipcPacket := scheduling_proto.IpcPacket{
		PacketType: &scheduling_proto.IpcPacket_InitPacket{},
	}

	bytes, err := protojson.Marshal(&ipcPacket)
	if err != nil {
		return fmt.Errorf("send offer IPC packet could not be marshaled: %w", err)
	}
	jsWorker.worker.Call("postMessage", string(bytes))
	return nil
}

// Send a message to the worker indicating to start
func (jsWorker *JsWorker) Terminate() {
	jsWorker.worker.Call("terminate")
}
