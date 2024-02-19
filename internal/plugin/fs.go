package plugin

import (
	"fmt"
	"io/fs"
	"os"
	"sync"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"

	"github.com/hack-pad/hackpadfs/mem"
)

var (
	onceAvailableFileSystem sync.Once
	availableFileSystems    map[config_proto.FileSystemType]func(any) (fs.FS, error)
)

type isRepositoryConfig_FileSystemConfig interface {
	isRepositoryConfig_FileSystemConfig()
}

// Singleton to allow multiple file systems implementation to register as an option
func AvailableFileSystems() map[config_proto.FileSystemType]func(any) (fs.FS, error) {
	onceAvailableFileSystem.Do(func() {
		availableFileSystems = map[config_proto.FileSystemType]func(any) (fs.FS, error){
			config_proto.FileSystemType_Os: func(config any) (fs.FS, error) {
				switch osConfig := config.(type) {
				case *config_proto.RepositoryConfig_Os:
					return os.DirFS(osConfig.Os.Directory), nil
				}

				return nil, fmt.Errorf("invalid config type passed")
			},
			config_proto.FileSystemType_Memory: func(config any) (fs.FS, error) {
				switch config.(type) {
				case *config_proto.RepositoryConfig_Memory:
					return mem.NewFS()
				}

				return nil, fmt.Errorf("invalid config type passed")
			},
		}
	})

	return availableFileSystems
}

// Get the file system associated to the given repository config
func GetFileSystem(repositoryConfig *config_proto.RepositoryConfig) (fs.FS, error) {
	var selectedFileSystem config_proto.FileSystemType

	switch repositoryConfig.FileSystemConfig.(type) {
	case *config_proto.RepositoryConfig_IndexedDb:
		selectedFileSystem = config_proto.FileSystemType_IndexedDb
	case *config_proto.RepositoryConfig_Memory:
		selectedFileSystem = config_proto.FileSystemType_Memory
	case *config_proto.RepositoryConfig_Os:
		selectedFileSystem = config_proto.FileSystemType_Os
	}

	fileSystemFactory, ok := AvailableFileSystems()[selectedFileSystem]
	if ok {
		return fileSystemFactory(repositoryConfig.FileSystemConfig)
	}

	return nil, fmt.Errorf("the requested file system type is not available in the current context")
}
