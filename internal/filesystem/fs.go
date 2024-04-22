package filesystem

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"sync"

	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	"github.com/hack-pad/hackpadfs/mem"
	"google.golang.org/protobuf/proto"
)

var (
	onceAvailableFileSystem    sync.Once
	availableFileSystems       map[filesystem_proto.FileSystemType]func(any) (fs.FS, error)
	registeredFileSystemsMutex sync.Mutex
	registeredFileSystems      map[string]fs.FS = map[string]fs.FS{}
)

type isRepositoryConfig_FileSystemConfig interface {
	isRepositoryConfig_FileSystemConfig()
}

// Singleton to allow multiple file systems implementation to register as an option
func AvailableFileSystems() map[filesystem_proto.FileSystemType]func(any) (fs.FS, error) {
	onceAvailableFileSystem.Do(func() {
		availableFileSystems = map[filesystem_proto.FileSystemType]func(any) (fs.FS, error){
			filesystem_proto.FileSystemType_Os: func(config any) (fs.FS, error) {
				switch osConfig := config.(type) {
				case *filesystem_proto.FileSystemSettings_Os:
					return os.DirFS(osConfig.Os.Directory), nil
				}

				return nil, fmt.Errorf("invalid config type passed")
			},
			filesystem_proto.FileSystemType_Memory: func(config any) (fs.FS, error) {
				switch config.(type) {
				case *filesystem_proto.FileSystemSettings_Memory:
					return mem.NewFS()
				case nil:
					return mem.NewFS()
				}

				return nil, fmt.Errorf("invalid config type passed")
			},
		}
	})

	return availableFileSystems
}

// Get the file system associated to the given repository config
func GetFileSystem(filesystemSettings *filesystem_proto.FileSystemSettings) (fs.FS, error) {
	var selectedFileSystem filesystem_proto.FileSystemType

	switch filesystemSettings.FileSystemSettings.(type) {
	case *filesystem_proto.FileSystemSettings_IndexedDb:
		selectedFileSystem = filesystem_proto.FileSystemType_IndexedDb
	case *filesystem_proto.FileSystemSettings_Memory:
		selectedFileSystem = filesystem_proto.FileSystemType_Memory
	case *filesystem_proto.FileSystemSettings_Os:
		selectedFileSystem = filesystem_proto.FileSystemType_Os
	}

	registeredFileSystemsMutex.Lock()
	defer registeredFileSystemsMutex.Unlock()

	// Hashing the config so we can get or create the appropriate file system
	b, err := proto.Marshal(filesystemSettings)
	if err != nil {
		return nil, fmt.Errorf("could not hash the given repository config: %w", err)
	}
	key := base64.StdEncoding.EncodeToString(b)
	if _, ok := registeredFileSystems[key]; ok {
		return registeredFileSystems[key], nil
	}

	fileSystemFactory, ok := AvailableFileSystems()[selectedFileSystem]
	if ok {
		fileSystem, err := fileSystemFactory(filesystemSettings.FileSystemSettings)
		registeredFileSystems[key] = fileSystem
		return fileSystem, err
	}

	return nil, fmt.Errorf("the requested file system type is not available in the current context")
}
