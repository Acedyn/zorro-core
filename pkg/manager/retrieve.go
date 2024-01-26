package manager

import (
	"sync"

	"github.com/Acedyn/zorro-core/internal/tools"
)

var (
	invokedTools []tools.Tool
	once         sync.Once
)

// Getter for the invoked tools singleton which holds the tools that have been invoked
func InvokedTools() []tools.Tool {
	once.Do(func() {
		invokedTools = []tools.Tool{}
	})

	return invokedTools
}

func RetrieveActions() []tools.Action {
	actions := []tools.Action{}

	return actions
}
