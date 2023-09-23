package processor

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/life4/genesis/maps"
  processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
)

// Wrapped processor with methods attached
type Processor struct {
  *processor_proto.Processor
}

// Start the client into a running client
func (processor *Processor) Start(
	metadata map[string]string,
	environ []string,
) (*PendingProcessor, error) {
	registration := make(chan error)
	pendingProcessor := &PendingProcessor{
		Processor:       processor,
		Registration: registration,
	}
	startingStatus := processor_proto.ProcessorStatus_STARTING
	pendingProcessor.Status = startingStatus
	pendingProcessor.Metadata = maps.Merge(pendingProcessor.GetMetadata(), metadata)

	// Build the command template
	template, err := template.New(processor.GetName()).Parse(pendingProcessor.GetStartProcessorTemplate())
	if err != nil {
		return nil, fmt.Errorf(
			"could not run client %s: Invalid launch template %w",
			pendingProcessor.GetName(),
			err,
		)
	}

	// Apply the metadata and the name on the template
	runCommand := &bytes.Buffer{}
	err = template.Execute(runCommand, struct {
		Name     string
		Label    string
		Version  string
		Metadata map[string]string
	}{
		Name:     pendingProcessor.GetName(),
		Label:    pendingProcessor.GetLabel(),
		Version:  pendingProcessor.GetVersion(),
		Metadata: pendingProcessor.GetMetadata(),
	})
	if err != nil {
		return nil, fmt.Errorf("could not run client %s: Templating error %w", pendingProcessor.GetName(), err)
	}

	// Build the subprocess's env with the context's environment variables
	splittedCommand := strings.Split(runCommand.String(), " ")
	clientCommand := exec.Command(splittedCommand[0], splittedCommand[1:]...)
	clientCommand.Env = environ

	// Start the subprocess
	err = clientCommand.Start()
	if err != nil {
		return nil, fmt.Errorf("an error occured while starting process for processor %s: %w", processor, err)
	}

	// Register the new client into the client queue and wait for it to be registered
	processorQueueLock.Lock()
	ProcessorQueue()[pendingProcessor.GetId()] = pendingProcessor
	processorQueueLock.Unlock()

	return pendingProcessor, <-registration
}

// Update the processor with a patch
func (processor *Processor) Patch(patch *Processor) bool {
	// Patch the local version of the client
	isPatched := false
	if maps.Equal(processor.Metadata, patch.GetMetadata()) {
		maps.Update(processor.Metadata, patch.GetMetadata())
		isPatched = true
	}
	if processor.GetLabel() != patch.GetLabel() {
		processor.Label = patch.GetLabel()
		isPatched = true
	}
	if processor.GetStatus() != patch.GetStatus() {
		processor.Status = patch.GetStatus()
		isPatched = true
	}

	return isPatched
}
