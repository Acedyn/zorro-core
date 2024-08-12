//go:build js && wasm

package wasm

import (
	"context"
	"syscall/js"
	"testing"
	"time"

	_ "github.com/Acedyn/zorro-core/testdata"

	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
)

var workerSignalingTest = &scheduling_proto.Processor{
	Name:    "wasm-processor-test",
	Version: "1.0",
	Paths: []*scheduling_proto.ProcessorPath{
		{
			FileSystem: &filesystem_proto.FileSystemSettings{
				FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
			},
			StartProcessorTemplate: "processors/wasm_worker_signaling.js",
		},
	},
}

// Start a web worker and make sure the communication protocol is correctly implemented
func OldTestWorkerSignaling(t *testing.T) {
	// Cancel the worker initialization after 2 seconds
	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelCtx()

	// Start the web worker
	jsWorker, err := NewWorker(workerSignalingTest)
	if err != nil {
		t.Errorf("An error occured when trying to start the processor worker: %s", err)
		return
	} else if jsWorker.worker.IsUndefined() {
		t.Errorf("Web worker instansiation failed")
	}

	// Make sure the worker is terminated at the end
	defer func() {
		jsWorker.Terminate()
		// We must wait for the worker to be terminated before the wasm context is dropped
		time.Sleep(time.Millisecond * 200)
	}()

	// Keeps track of the reveived messages from the worker
	messages := make(chan js.Value)

	// Patch the onmessage to inject testing logic
	jsWorker.worker.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 1 {
			messages <- args[0]
		}

		return jsWorker.onWorkerMessage(this, args)
	}))

	// We expect to receive a specific sequence of message from the worker. If it doesn't respect
	// this order we conclude that something went wrong
	expectedMessageProcedures := []string{"FILE_QUERY", "FILE_QUERY"}

	// Initialize the communication using the web worker protocol
	jsWorker.Init()

	for i := 0; i < len(expectedMessageProcedures); i += 1 {
		select {
		case response := <-messages:
			// Make sure the response is valid
			if response.IsUndefined() || response.Get("data").Get("procedure").IsUndefined() {
				t.Errorf("Processor web worker returned undefined or invalid object as a response (%s)", response.String())
				return
			}
			// Test if we received the expected response, wich ensure that everything went wright on the worker's side
			if response.Get("data").Get("procedure").String() != expectedMessageProcedures[i] {
				t.Errorf("Processor web worker returned %s as a response where %s was expected",
					response.Get("data").Get("procedure").String(),
					expectedMessageProcedures[i],
				)
				return
			}
		case <-ctx.Done():
			t.Errorf("No message received from web worker: context timed out")
			return
		}
	}
}
