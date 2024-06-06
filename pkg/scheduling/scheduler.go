package scheduling

import (
	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/pkg/context"
)

// A scheduler will manage the processors that will execute the
// tools. It is responsible for starting and stopping schedulers when needed.
type Scheduler interface {
	GetOrStartProcessor(*context.Context, *processor.ProcessorQuery) (*processor.RegisteredProcessor, error)
}
