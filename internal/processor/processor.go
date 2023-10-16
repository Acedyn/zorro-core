package processor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Acedyn/zorro-core/internal/utils"

	processor_proto "github.com/Acedyn/zorro-proto/zorroprotos/processor"
	"github.com/google/uuid"
	"github.com/hoisie/mustache"
	"github.com/life4/genesis/maps"
)

// Wrapped processor with methods attached
type Processor struct {
	*processor_proto.Processor
}

// Start the client into a running client. This methods make a copy of the processor
func (processor Processor) Start(
	metadata map[string]string,
	environ []string,
	commandPaths []string,
) (*PendingProcessor, error) {
	registration := make(chan error)
	pendingProcessor := &PendingProcessor{
		Processor:    &processor,
		Registration: registration,
	}
	startingStatus := processor_proto.ProcessorStatus_STARTING
	pendingProcessor.Status = startingStatus
	pendingProcessor.Metadata = maps.Merge(processor.GetMetadata(), metadata)

	// Generate an ID for this new processor
	processor.Id = uuid.New().String()

	// Apply the metadata and the name on the template
	runCommand, err := processor.buildCommand(commandPaths)
	if err != nil {
		return nil, fmt.Errorf("could not run processor (%s): %w", processor.GetName(), err)
	}

	// Build the subprocess's env with the context's environment variables
	splittedCommand := strings.Split(runCommand, " ")
	processorCommand := exec.Command(splittedCommand[0], splittedCommand[1:]...)
	processorCommand.Env = environ

	processorCommand.Stdout = io.MultiWriter(
		&pendingProcessor.Stdout,
		utils.NewPrefixedWriter(os.Stdout, "[PROCESSOR: "+processor.Id+"] "),
	)
	processorCommand.Stderr = io.MultiWriter(
		&pendingProcessor.Stderr,
		utils.NewPrefixedWriter(os.Stderr, "[PROCESSOR: "+processor.Id+"] "),
	)

	// Start the subprocess
	err = processorCommand.Start()
	if err != nil {
		return nil, fmt.Errorf("an error occured while starting process for processor (%s) with command %s: %w", processor, splittedCommand, err)
	}

	// Register the new client into the client queue and wait for it to be registered
	processorQueueLock.Lock()
	ProcessorQueue()[pendingProcessor.GetId()] = pendingProcessor
	processorQueueLock.Unlock()

	// Wait for the command to end so we can get the output code
	commandResult := make(chan error)
	go func() {
		output := processorCommand.Wait()
		if output != nil {
			commandResult <- fmt.Errorf("the processor command %s exited: %w", splittedCommand, output)
		} else {
			commandResult <- nil
		}
	}()

	// We wait for either the processor to be registered or the command to error out
	err = nil
	select {
	case registrationOutput := <-registration:
		err = registrationOutput
	case commandOutput := <-commandResult:
		err = commandOutput
	}

	return pendingProcessor, err
}

// Build the command used to start the processor
func (processor *Processor) buildCommand(commandsPaths []string) (string, error) {
	// Build the command template
	template, err := mustache.ParseString(processor.GetStartProcessorTemplate())
	if err != nil {
		return "", fmt.Errorf(
			"could not parse launch template %w",
			err,
		)
	}

	// Render the template to the actual string
	return template.Render(map[string]any{
		"name":     processor.GetName(),
		"label":    processor.GetLabel(),
		"version":  processor.GetVersion(),
		"id":       processor.GetId(),
		"metadata": processor.GetMetadata(),
		"commands": commandsPaths,
	}), nil
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
