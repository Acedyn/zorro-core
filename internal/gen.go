//go:generate protoc --go_out=. --go_opt=paths=source_relative ./context/context.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative ./context/plugin.proto

//go:generate protoc --go_out=. --go_opt=paths=source_relative ./config/config.proto

//go:generate protoc --go_out=. --go_opt=paths=source_relative ./tools/tool.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative ./tools/action.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative ./tools/command.proto

//go:generate protoc --go_out=. --go_opt=paths=source_relative ./scheduling/client.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative ./scheduling/scheduler.proto
package internal
