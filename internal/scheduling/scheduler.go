package scheduling

import (
	"sync"

	"github.com/Acedyn/zorro-core/internal/tools"
)

var (
	availableSchedulers     map[string]Scheduler
	onceAvailableSchedulers sync.Once
)

type SchedulerInfo struct {
	Name string
}

type Scheduler interface {
	// Start the evantual scheduler listenners
	Initialize()
	// Identifiers used to match againts the scheduler query
	GetInfo() SchedulerInfo
	// Request to the scheduler to execute the command query
	ScheduleCommand(*tools.CommandQuery)
}

// Getter for the available schedulers singleton
func AvailableSchedulers() map[string]Scheduler {
	onceAvailableSchedulers.Do(func() {
		availableSchedulers = map[string]Scheduler{}
	})

	return availableSchedulers
}

// Listen for the command queue's queries and schedule it to the appropriate scheduler
func InitializeAvailableSchedulers() {
	for _, scheduler := range AvailableSchedulers() {
		scheduler.Initialize()
	}
}

// Listen for the command queue's queries and schedule it to the appropriate scheduler
func ListenCommandQueries() {
	for commandQuery := range tools.CommandQueue() {
		// For now we only schedule commands with the suprocess scheduler
		AvailableSchedulers()["subprocess"].ScheduleCommand(commandQuery)
	}
}
