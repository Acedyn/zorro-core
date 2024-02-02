package subprocess

import (
	"github.com/Acedyn/zorro-core/internal/processor"

	scheduling_proto "github.com/Acedyn/zorro-proto/zorroprotos/scheduling"
	"github.com/life4/genesis/slices"
)

// Wrapped processor query with methods attached
type ProcessorQuery struct {
	*scheduling_proto.ProcessorQuery
}

// Test if a client matches the query's requirements
func (query *ProcessorQuery) MatchProcessor(processor *processor.Processor) bool {
	// Test the name
	if query.Name != nil {
		// Some clients are supersets of other clients
		// If so they should match also their subsets
		subsets := append(processor.GetSubsets(), processor.GetName())
		if !slices.Contains(subsets, query.GetName()) {
			return false
		}
	}
	// Test the ID
	if query.Id != nil {
		if query.GetId() != processor.GetId() {
			return false
		}
	}
	// Test the Metadata
	for key, value := range query.GetMetadata() {
		metadata, ok := processor.GetMetadata()[key]
		if !ok || metadata != value {
			return false
		}
	}
	return true
}
