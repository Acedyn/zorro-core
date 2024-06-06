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

var startProcessorTest = &scheduling_proto.Processor{
	Name:    "wasm-processor-test",
	Version: "1.0",
	Paths: []*scheduling_proto.ProcessorPath{
		{
			FileSystem: &filesystem_proto.FileSystemSettings{
				FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
			},
			StartProcessorTemplate: "processors/wasm_worker.js",
			StartProgramTemplate:   "processors/wasm_worker.js",
		},
	},
}

func TestStartProcessorWorker(t *testing.T) {
	// Cancel the worker initialization after 2 seconds
	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelCtx()

	// Start the web worker
	worker, err := startProcessorWorker(startProcessorTest)
	if err != nil {
		t.Errorf("An error occured when trying to start the processor worker: %s", err)
		return
	} else if worker.IsUndefined() {
		t.Errorf("Web worker instansiation failed")
	}

	// Try to communicate with the web worker
	workerMessageChan := make(chan error)
	worker.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			workerMessageChan <- fmt.Errorf("received message from worker with %d argument instread of 1", len(args))
			return nil
		}
		if args[0].Get("data").String() != "pong" {
			workerMessageChan <- fmt.Errorf("message received from worker invalid: (%s instead of hello world)", args[0].Get("data").String())
			return nil
		}

		workerMessageChan <- nil
		return nil
	}))

	// Initialize the communication
	worker.Call("postMessage", "ping")

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
