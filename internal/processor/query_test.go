package processor

import (
	"testing"

	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
)

var processorQuery = ProcessorQuery{
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
	testProcessorA := scheduling_proto.Processor{
		Name:    "foo",
		Version: "0.2.3",
		Metadata: map[string]string{
			"a": "one",
		},
	}
	if processorQuery.MatchProcessor(&testProcessorA) {
		t.Errorf("Invalid match: %s should not match %s", processorQuery.String(), testProcessorA.String())
	}

	testProcessorB := scheduling_proto.Processor{
		Name:    "foo",
		Version: "0.2.3",
		Metadata: map[string]string{
			"a": "one",
			"b": "two",
		},
	}
	if !processorQuery.MatchProcessor(&testProcessorB) {
		t.Errorf("Invalid match: %s should match %s", processorQuery.String(), testProcessorA.String())
	}
}
