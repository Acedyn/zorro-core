//go:build js && wasm

package wasm

import (
	"context"
	"fmt"

	"github.com/Acedyn/zorro-core/internal/network/rpc"
	"github.com/Acedyn/zorro-core/internal/network/wrtc"
	"github.com/Acedyn/zorro-core/internal/processor"
	zorro_context "github.com/Acedyn/zorro-core/pkg/context"

	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
)

type WasmScheduler struct {
	pool *processor.ProcessorPool
}

func (scheduler *WasmScheduler) GetOrStartProcessor(ctx context.Context, zorro_context *zorro_context.Context, query *processor.ProcessorQuery) (*processor.RegisteredProcessor, error) {
	// First find a potential running processors that matches the query
	if registeredClient := scheduler.pool.FindRegisteredProcessor(query); registeredClient != nil {
		return registeredClient, nil
	}

	// Find and start the processor
	processorToStart, err := findProcessor(zorro_context, query)
	if err != nil {
		return nil, fmt.Errorf("could not find processor")
	}
	jsWorker, err := NewWorker(processorToStart)
	if err != nil {
		return nil, fmt.Errorf("an error occured when starting processor web worker: %w", err)
	} else if jsWorker.worker.IsUndefined() {
		return nil, fmt.Errorf("web worker instansiation failed")
	}

	// Connect with the web worker via wrtc
	connection, err := setupWrtcConnection(ctx, jsWorker)
	if err != nil {
		return nil, fmt.Errorf("could not connect with processor via wrtc: %w", err)
	}
	// Create a client able to send grpc requests to the web worker
	grpcClient, err := rpc.NewClient(connection, rpc.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("could initialize grpc client via wrtc: %w", err)
	}

	// Register the processor so it can be used by further requests
	registeredProcessor := scheduler.pool.RegisterProcessor(processorToStart, grpcClient)
	return registeredProcessor, nil
}

// Find a processor from the given context that will satisfy the processor query
func findProcessor(zorro_context *zorro_context.Context, query *processor.ProcessorQuery) (*scheduling_proto.Processor, error) {
	for _, availableProcessor := range zorro_context.AvailableProcessors() {
		if query.MatchProcessor(availableProcessor) {
			return availableProcessor, nil
		}
	}

	return nil, fmt.Errorf("no available processors matched the query %+v", query)
}

// Because of a limitation in the wrtc's implementation in most browsers, web workers
// can't create or use WRTC connections. The temporary solution to this issue is that
// the WRTC connection is created an used in the main window and forward the requests
// and answers with the worker using postMessage. This function is temporary and will
// be replaced by the workerSignaler when the support for transferable datachannels
// is better.
// Chromium: https://issues.chromium.org/issues/40787712
// Safary: OK
// Firefox: ?
func setupWrtcConnection(ctx context.Context, jsWorker *JsWorker) (*wrtc.WRTCConnection, error) {
	tempWorkerSignaler := newWorkerTempSignaler()

	// Group the output of the connection's initialisation
	type wrtcConnectionOutput struct {
		*wrtc.WRTCConnection
		error
	}
	workerConnectionChan := make(chan wrtcConnectionOutput)
	processorConnectionChan := make(chan wrtcConnectionOutput)

	// Connect the two peers together in their own goroutines
	go func() {
		workerConnection, err := wrtc.NewConnectionAsOfferer(
			wrtc.WithSignalerWRTCOption(tempWorkerSignaler),
			wrtc.WithContextWRTCOption(ctx),
		)
		workerConnectionChan <- wrtcConnectionOutput{
			WRTCConnection: workerConnection,
			error:          err,
		}
	}()
	go func() {
		processorConnection, err := wrtc.NewConnectionAsOfferer(
			wrtc.WithSignalerWRTCOption(tempWorkerSignaler),
			wrtc.WithContextWRTCOption(ctx),
		)
		processorConnectionChan <- wrtcConnectionOutput{
			WRTCConnection: processorConnection,
			error:          err,
		}
	}()

	// Get the worker's established connection
	select {
	case workerConnection := <-workerConnectionChan:
		if workerConnection.error != nil {
			return nil, fmt.Errorf("wasm worker's connection error: %w", workerConnection.error)
		}
		// Set the worker's connection and handle the pending forward requests
		jsWorker.Connection = workerConnection.WRTCConnection
		jsWorker.HandlePendingForward()
	case <-ctx.Done():
		return nil, fmt.Errorf("wasm worker's wrtc connection canceled")
	}

	// Get the processor's established connection
	select {
	case processorConnection := <-processorConnectionChan:
		if processorConnection.error != nil {
			return nil, fmt.Errorf("wasm processor's connection error: %w", processorConnection.error)
		}
		return processorConnection.WRTCConnection, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("wasm processor's wrtc connection canceled")
	}
}
