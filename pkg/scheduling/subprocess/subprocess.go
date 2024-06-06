package subprocess

import (
	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/pkg/context"
)

type SubprocessScheduler struct {
	pool *processor.ProcessorPool
}

func (scheduler *SubprocessScheduler) GetOrStartProcessor(c *context.Context, query *processor.ProcessorQuery) (*processor.RegisteredProcessor, error) {
	// First find a potential running processors that matches the query
	if registeredClient := scheduler.pool.FindRegisteredProcessor(query); registeredClient != nil {
		return registeredClient, nil
	}

	return nil, nil
}
