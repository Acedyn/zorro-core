package processor

import (
	"runtime"
	"testing"
	"time"

	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
	"github.com/life4/genesis/maps"
)

// Mocked processors to start
var startClientTestWindows = Processor{
	Processor: &processor_proto.Processor{
		Name:                   "cmd",
		StartProcessorTemplate: "{{.Name}}",
	},
}

var startClientTestLinux = Processor{
	Processor: &processor_proto.Processor{
		Name:                   "bash",
		StartProcessorTemplate: "{{.Name}}",
	},
}

// Mocked scheduler that will falsely register the processors
func mockedScheduler(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		default:
			processorQueueLock.Lock()
			pendingProcessors := maps.Values(ProcessorQueue())
			processorQueueLock.Unlock()
			for _, pendingProcessor := range pendingProcessors {
				UnQueueProcessor(pendingProcessor)
				pendingProcessor.Registration <- nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Start the Start() methods of a processor
func TestStartProcessor(t *testing.T) {
	stopScheduler := make(chan bool)
	defer func() { stopScheduler <- true }()
	go mockedScheduler(stopScheduler)

	var startProcessorTest *Processor = nil
	switch runtime.GOOS {
	case "windows":
		startProcessorTest = &startClientTestWindows
	case "linux":
		startProcessorTest = &startClientTestLinux
	default:
		startProcessorTest = &startClientTestLinux
	}

	_, err := startProcessorTest.Start(map[string]string{}, []string{})
	if err != nil {
		t.Errorf("An error occured while running client %s: %s", startProcessorTest, err.Error())
		return
	}
}
