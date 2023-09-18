//go:generate protoc --go_out=. --go_opt=paths=source_relative ./internal/context/context.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative ./internal/context/plugin.proto

//go:generate protoc --go_out=. --go_opt=paths=source_relative ./internal/config/config.proto

//go:generate protoc --go_out=. --go_opt=paths=source_relative ./internal/tools/tool.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative ./internal/tools/action.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative ./internal/tools/command.proto

//go:generate protoc --go_out=. --go_opt=paths=source_relative ./internal/scheduling/client.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./internal/scheduling/scheduler.proto
package internal
