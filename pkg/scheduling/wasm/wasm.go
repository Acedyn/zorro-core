//go:build js && wasm

package wasm

import (
	"fmt"

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
	jsWorker, err := NewWorker(processorToStart)
	if err != nil {
		return nil, fmt.Errorf("an error occured when starting processor web worker: %w", err)
	} else if jsWorker.worker.IsUndefined() {
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
