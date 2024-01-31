package scheduling_test

import (
	"testing"

	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/internal/scheduling"

	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
)

var processorQuery = scheduling.ProcessorQuery{
	ProcessorQuery: &scheduling_proto.ProcessorQuery{
		Name:    &[]string{"foo"}[0],
		Version: &[]string{"0.2.3"}[0],
		Metadata: map[string]string{
			"a": "one",
			"b": "two",
		},
	},
}

func TestProcessorQuery(t *testing.T) {
	testProcessorA := processor.Processor{
		Processor: &processor_proto.Processor{
			Name:    "foo",
			Version: "0.2.3",
			Metadata: map[string]string{
				"a": "one",
			},
		},
	}
	if processorQuery.MatchProcessor(&testProcessorA) {
		t.Errorf("Invalid match: %s should not match %s", processorQuery, testProcessorA)
	}

	testProcessorB := processor.Processor{
		Processor: &processor_proto.Processor{
			Name:    "foo",
			Version: "0.2.3",
			Metadata: map[string]string{
				"a": "one",
				"b": "two",
			},
		},
	}
	if !processorQuery.MatchProcessor(&testProcessorB) {
		t.Errorf("Invalid match: %s should match %s", processorQuery, testProcessorA)
	}
}
