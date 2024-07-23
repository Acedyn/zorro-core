//go:build js && wasm

package wasm

import (
	"fmt"
	"io"
	"strings"
	"syscall/js"

	"github.com/Acedyn/zorro-core/internal/filesystem"
	"github.com/Acedyn/zorro-core/internal/utils"

	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
	"github.com/life4/genesis/slices"
)

type JsWorker struct {
	workerApi js.Value
	blobApi   js.Value
	urlApi    js.Value
	worker    js.Value
	processor *scheduling_proto.Processor
	path      *scheduling_proto.ProcessorPath
}

// Start a new web worker that will run the given processor's script
func NewWorker(processorToStart *scheduling_proto.Processor) (*JsWorker, error) {
	// Initialize the worker struct
	jsWorker := JsWorker{
		processor: processorToStart,
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

		// Store the file content in a blob
		processorScriptFile := jsWorker.blobApi.New(
			[]any{string(fileData)},
			map[string]any{"type": "text/javascript"},
		)

		processorScriptUrl := jsWorker.urlApi.Call("createObjectURL", processorScriptFile)
		jsWorker.worker = jsWorker.workerApi.New(processorScriptUrl)
		break
	}

	// If none of the proposed script path worked return an error
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
		jsWorker.logError("Received message from web worker with %d argument instread of 1")
		return nil
	}

	// The wasm worker for a processor must follow a strict protocol defined here
	messageData := args[0].Get("data")
	procedure := messageData.Get("procedure").String()
	arguments := messageData.Get("args")
	switch procedure {
	case "FILE_QUERY":
		err := jsWorker.onFileQueryMessage(arguments)
		if err != nil {
			jsWorker.logError("Could not satisfy file query (%s)", err.Error())
		}
	case "ERROR":
		jsWorker.onErrorMessage(arguments)
	case "SDP_OFFER":
		jsWorker.onSdpOffer(arguments)
	default:
		jsWorker.logError("Received unknown message procedure (%s)", procedure)
	}

	return nil
}

// Handler for messages comming from the worker with FILE_QUERY as procedure
func (jsWorker *JsWorker) onFileQueryMessage(messageData js.Value) error {
	// Make sure something is always returned
	filePath := messageData.Get("path")

	if filePath.IsUndefined() {
		return fmt.Errorf("missing key 'path' in file query message")
	}

	processorScriptFile := map[string]any{
		"procedure": "FILE_RESPONSE",
		"args": map[string]any{
			"path": filePath,
		},
	}
	defer jsWorker.worker.Call("postMessage", processorScriptFile)

	processorFileSystem, err := filesystem.GetFileSystem(jsWorker.path.GetFileSystem())
	if err != nil {
		return err
	}

	// Read the file
	fileHandle, err := processorFileSystem.Open(filePath.String())
	if err != nil {
		return err
	}
	defer fileHandle.Close()
	fileData, err := io.ReadAll(fileHandle)
	if err != nil {
		return err
	}

	// Store the file content in a blob
	processorScriptFile["args"] = map[string]any{
		"content": jsWorker.blobApi.New([]any{string(fileData)}),
		"path":    filePath,
	}
	return nil
}

// Handler for messages comming from the worker with ERROR as procedure
func (jsWorker *JsWorker) onErrorMessage(messageData js.Value) {
	jsWorker.logError(messageData.Get("exception").String())
}

// Handler for messages comming from the worker with SDP_OFFER as procedure
func (jsWorker *JsWorker) onSdpOffer(messageData js.Value) {

}

// Send a message to the worker indicating to start
func (jsWorker *JsWorker) Init() {
	jsWorker.worker.Call("postMessage", map[string]any{"procedure": "INIT"})
}

// Send a message to the worker indicating to start
func (jsWorker *JsWorker) Terminate() {
	jsWorker.worker.Call("terminate")
}
