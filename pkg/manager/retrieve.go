package manager

import (
	"sync"

	"github.com/Acedyn/zorro-core/internal/tools"
)

var (
	invokedActions  []*tools.Action
	onceActions     sync.Once
	invokedCommands []*tools.Command
	onceCommands    sync.Once
)

// Getter for the invoked commands singleton
func InvokedCommands() []*tools.Command {
	onceCommands.Do(func() {
		invokedCommands = []*tools.Command{}
	})

	return invokedCommands
}

// Getter for the invoked actions singleton
func InvokedActions() []*tools.Action {
	onceActions.Do(func() {
		invokedActions = []*tools.Action{}
	})

	return invokedActions
}
