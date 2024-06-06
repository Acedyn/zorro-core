//go:build js && wasm

package wasm

import (
	"fmt"
	"io"
	"syscall/js"

	"github.com/Acedyn/zorro-core/internal/filesystem"
	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/pkg/context"

	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
)

type WasmScheduler struct {
	pool *processor.ProcessorPool
}

func (scheduler *WasmScheduler) GetOrStartProcessor(c *context.Context, query *processor.ProcessorQuery) (*processor.RegisteredProcessor, error) {
	// First find a potential running processors that matches the query
	if registeredClient := scheduler.pool.FindRegisteredProcessor(query); registeredClient != nil {
		return registeredClient, nil
	}

	// Find and start the processor
	processorToStart, err := findProcessor(c, query)
	if err != nil {
		return nil, fmt.Errorf("could not find processor")
	}
	worker, err := startProcessorWorker(processorToStart)
	if err != nil {
		return nil, fmt.Errorf("an error occured when starting processor web worker: %w", err)
	} else if worker.IsUndefined() {
		return nil, fmt.Errorf("web worker instansiation failed")
	}

	return nil, nil
}

// Find a processor from the given context that will satisfy the processor query
func findProcessor(c *context.Context, query *processor.ProcessorQuery) (*scheduling_proto.Processor, error) {
	for _, availableProcessor := range c.AvailableProcessors() {
		if query.MatchProcessor(availableProcessor) {
			return availableProcessor, nil
		}
	}

	return nil, fmt.Errorf("no available processors matched the query %+v", query)
}

// Start a new web worker that will run the given processor's script
func startProcessorWorker(processorToStart *scheduling_proto.Processor) (js.Value, error) {
	// Get and check the worker API
	jsWorker := js.Global().Get("Worker")
	if jsWorker.IsUndefined() {
		return js.ValueOf(nil), fmt.Errorf("web worker api not supported")
	}

	// Get and check the blob API
	jsBlob := js.Global().Get("Blob")
	if jsBlob.IsUndefined() {
		return js.ValueOf(nil), fmt.Errorf("blob api not supported")
	}

	// Get and check the url API
	jsUrl := js.Global().Get("URL")
	if jsUrl.IsUndefined() {
		return js.ValueOf(nil), fmt.Errorf("url api not supported")
	}

	// Start a worker that will execute the processor's script
	var processorWorker js.Value
	errors := []error{}
	// Try every configured path to start the processor
	for _, processorPath := range processorToStart.GetPaths() {
		processorFileSystem, err := filesystem.GetFileSystem(processorPath.GetFileSystem())
		if err != nil {
			errors = append(errors, err)
			continue
		}

		fileHandle, err := processorFileSystem.Open(processorPath.GetStartProcessorTemplate())
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

		processorScriptFile := jsBlob.New(
			[]any{string(fileData)},
			map[string]any{"type": "text/javascript"},
		)

		processorScriptUrl := jsUrl.Call("createObjectURL", processorScriptFile)
		processorWorker = jsWorker.New(processorScriptUrl)
		break
	}

	if processorWorker.IsUndefined() {
		return js.ValueOf(nil), fmt.Errorf(
			"no valid processor could be started as a web worker (%w)",
			fmt.Errorf(""),
		)
	}

	return processorWorker, nil
}
