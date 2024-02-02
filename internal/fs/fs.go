package fs

import (
	"fmt"
	"sync"

	"github.com/Acedyn/zorro-core/internal/utils"
	"github.com/Acedyn/zorro-core/pkg/config"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	"github.com/hack-pad/hackpadfs"
)

var (
	onceAvailableFileSystem sync.Once
	availableFileSystems    map[config_proto.FileSystemType]func(any) (hackpadfs.FS, error)
	onceFileSystems         sync.Once
	fileSystems             map[string]hackpadfs.FS
)

// Singleton to allow multiple file systems implementation to register as an option
func AvailableFileSystems() map[config_proto.FileSystemType]func(any) (hackpadfs.FS, error) {
	onceAvailableFileSystem.Do(func() {
		availableFileSystems = map[config_proto.FileSystemType]func(any) (hackpadfs.FS, error){}
	})

	return availableFileSystems
}

// List of file systems to iterate over, following the user configuration
func FileSystems() map[string]hackpadfs.FS {
	globalConfig := config.GlobalConfig()
	registedFileSystems := AvailableFileSystems()
	fileSystems := map[string]hackpadfs.FS{}

	for name, fileSystemConfig := range globalConfig.GetFileSystemsConfig().GetFileSystems() {
		fileSystemImplementation, ok := registedFileSystems[fileSystemConfig.Type]
		if !ok {
			utils.Logger().Warn(fmt.Sprintf("Could not load file system of type %d: Implementation missing in current build", fileSystemConfig.Type))
			continue
		}

		fileSystem, err := fileSystemImplementation(fileSystemConfig.GetConfig())
		if err != nil {
			utils.Logger().Error(fmt.Sprintf("Could not load file system of type %d: %s", fileSystemConfig.Type, err.Error()))
			continue
		}
		fileSystems[name] = fileSystem
	}

	return fileSystems
}
