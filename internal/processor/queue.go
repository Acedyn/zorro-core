package processor

import (
	"sync"

  processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
)

var (
	processorQueueLock = &sync.Mutex{}
	processorQueue     map[string]*PendingProcessor
	once            sync.Once
)

// Processor that is waiting to be registered
type PendingProcessor struct {
	*Processor
	Registration chan error
}

// Getter for the processor queue singleton which holds the queue
// of client processors to be registered
func ProcessorQueue() map[string]*PendingProcessor {
	once.Do(func() {
		processorQueue = map[string]*PendingProcessor{}
	})

	return processorQueue
}

// Check if the processor was queued and remove it from the queue
func UnQueueProcessor(processor *processor_proto.ProcessorQuery) *PendingProcessor {
	processorQueueLock.Lock()
	defer processorQueueLock.Unlock()

	if clientHandle, ok := ProcessorQueue()[processor.GetId()]; ok {
		delete(ProcessorQueue(), processor.GetId())
		return clientHandle
	}

	return nil
}

