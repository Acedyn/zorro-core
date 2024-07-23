//go:build js && wasm

package wasm

import (
	"testing"

	_ "github.com/Acedyn/zorro-core/testdata"

	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	scheduling_proto "github.com/Acedyn/zorro-proto/gen/protos/scheduling"
)

var startProcessorTest = &scheduling_proto.Processor{
	Name:    "wasm-processor-test",
	Version: "1.0",
	Paths: []*scheduling_proto.ProcessorPath{
		{
			FileSystem: &filesystem_proto.FileSystemSettings{
				FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
			},
			StartProcessorTemplate: "processors/go_processor.js",
		},
	},
}

func TestStartWasmProcessor(t *testing.T) {
}
