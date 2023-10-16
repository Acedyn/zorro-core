package scheduling

import (
	"testing"
	"time"

	"github.com/Acedyn/zorro-core/internal/context"
	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/life4/genesis/maps"

	context_proto "github.com/Acedyn/zorro-proto/zorroprotos/context"
	plugin_proto "github.com/Acedyn/zorro-proto/zorroprotos/plugin"
	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
)

// Mocked context to test processor queries
var contextTest = context.Context{
	Context: &context_proto.Context{
		Plugins: []*plugin_proto.Plugin{
			{
				Processors: []*processor_proto.Processor{
					{
						Name:                   "bash",
						StartProcessorTemplate: "{{name}}",
					},
					{
						Name:                   "cmd",
						StartProcessorTemplate: "{{name}}",
					},
				},
			},
		},
	},
}

// Mocked processor queries
var processorQueryTests = []*ProcessorQuery{
	{
		ProcessorQuery: &scheduling_proto.ProcessorQuery{
			Name:    &[]string{"bash"}[0],
			Context: contextTest.Context,
		},
	},
	{
		ProcessorQuery: &scheduling_proto.ProcessorQuery{
			Name:    &[]string{"foo"}[0],
			Version: &[]string{"2.3"}[0],
		},
	},
}

// Mocked processor pool to fake a list of registered processors
var runningProcessorPool = map[string]*RegisteredProcessor{
	"": {
		Processor: &processor.Processor{
			Processor: &processor_proto.Processor{
				Name:    "foo",
				Version: "2.3",
			},
		},
	},
}

// Mocked scheduler that will falsely register the processors
func mockedScheduler(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		default:
			pendingProcessors := maps.Values(processor.ProcessorQueue())
			for _, pendingProcessor := range pendingProcessors {
				registerProcessor(pendingProcessor.Processor, "", nil)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Test the GetOrStartProcessor function
func TestGetOrStartProcessor(t *testing.T) {
	stopScheduler := make(chan bool)
	defer func() { stopScheduler <- true }()
	go mockedScheduler(stopScheduler)

	for processorId, runningProcessor := range runningProcessorPool {
		processorPoolLock.Lock()
		ProcessorPool()[processorId] = runningProcessor
		processorPoolLock.Unlock()
	}

	for _, processorQueryTest := range processorQueryTests {
		_, err := GetOrStartProcessor(processorQueryTest)
		if err != nil {
			t.Errorf("An error occured while getting client from query %s: %s", processorQueryTest, err.Error())
			return
		}
	}
}
