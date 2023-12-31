package processor

import (
	"bytes"
	"sync"
)

var (
	processorQueueLock = &sync.Mutex{}
	processorQueue     map[string]*PendingProcessor
	once               sync.Once
)

// Processor that is waiting to be registered
type PendingProcessor struct {
	*Processor
	Registration chan error
	Stdout       bytes.Buffer
	Stderr       bytes.Buffer
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
func UnQueueProcessor(processorId string) *PendingProcessor {
	processorQueueLock.Lock()
	defer processorQueueLock.Unlock()

	if pendingProcessor, ok := ProcessorQueue()[processorId]; ok {
		delete(ProcessorQueue(), processorId)
		return pendingProcessor
	}

	return nil
}
