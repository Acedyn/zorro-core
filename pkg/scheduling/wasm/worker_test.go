//go:build js && wasm

package wasm

import (
	"context"
	"fmt"
	"syscall/js"
	"testing"
	"time"

	_ "github.com/Acedyn/zorro-core/testdata"

	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
)

var startProcessorWorkerTest = &scheduling_proto.Processor{
	Name:    "wasm-processor-test",
	Version: "1.0",
	Paths: []*scheduling_proto.ProcessorPath{
		{
			FileSystem: &filesystem_proto.FileSystemSettings{
				FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
			},
			StartProcessorTemplate: "processors/wasm_worker.js",
		},
	},
}

// Test starting a web worker and communicate with it via simple sendMessage and postMessage
func TestStartProcessorWorker(t *testing.T) {
	// Cancel the worker initialization after 2 seconds
	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelCtx()

	// Start the web worker
	jsWorker, err := NewWorker(startProcessorWorkerTest)
	if err != nil {
		t.Errorf("An error occured when trying to start the processor worker: %s", err)
		return
	} else if jsWorker.worker.IsUndefined() {
		t.Errorf("Web worker instansiation failed")
	}

	// Try to communicate with the web worker
	workerMessageChan := make(chan error)
	jsWorker.worker.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			workerMessageChan <- fmt.Errorf(
				"received message from worker with %d argument instread of 1",
				len(args),
			)
			return nil
		}
		if args[0].Get("data").String() != "pong" {
			workerMessageChan <- fmt.Errorf(
				"message received from worker invalid: (%s instead of hello world)",
				args[0].Get("data").String(),
			)
			return nil
		}

		workerMessageChan <- nil
		return nil
	}))

	// Initialize the communication
	jsWorker.worker.Call("postMessage", "ping")

	select {
	case err := <-workerMessageChan:
		if err != nil {
			t.Errorf("An error occured when communicating with the processor's web worker: %s", err)
		}
		return
	case <-ctx.Done():
		t.Errorf("No message received from web worker: context timed out")
		return
	}
}

var workerProtocolTest = &scheduling_proto.Processor{
	Name:    "wasm-processor-test",
	Version: "1.0",
	Paths: []*scheduling_proto.ProcessorPath{
		{
			FileSystem: &filesystem_proto.FileSystemSettings{
				FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
			},
			StartProcessorTemplate: "processors/wasm_worker_protocol.js",
		},
	},
}

// Start a web worker and make sure the communication protocol is correctly implemented
func TestWorkerProtocol(t *testing.T) {
	// Cancel the worker initialization after 2 seconds
	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelCtx()

	// Start the web worker
	jsWorker, err := NewWorker(workerProtocolTest)
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
