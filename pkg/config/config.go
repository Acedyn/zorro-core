package config

import (
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/Acedyn/zorro-core/internal/filesystem"

	config_proto "github.com/Acedyn/zorro-proto/gen/protos/config"
	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	config            *config_proto.Config
	once              sync.Once
	globalConfigMutex sync.Mutex
)

// Getter for the config singleton
func GlobalConfig() *config_proto.Config {
	once.Do(func() {
		// TODO: Get the config from somewhere on the computer
		config = DEFAULT
	})

	return config
}

// Load the global config and apply the config file content
// WARNING: This function is not threadsafe, make sure to use it during initialization of the program
func LoadConfig(path string, filesystemSettings *filesystem_proto.FileSystemSettings) (*config_proto.Config, error) {
	config = GlobalConfig()

	selectedFs, err := filesystem.GetFileSystem(filesystemSettings)
	if err != nil {
		return nil, fmt.Errorf("invalid file system (%s): %w", filesystemSettings, err)
	}

	fileHandle, err := selectedFs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file (%s): %w", path, err)
	}
	defer fileHandle.Close()

	// Parse the plugin data
	fileData, err := io.ReadAll(fileHandle)
	if err != nil {
		return nil, fmt.Errorf("could not read config file (%s): %w", path, err)
	}

	switch filepath.Ext(path) {
	case ".json":
		err := protojson.Unmarshal(fileData, config)
		if err != nil {
			return nil, fmt.Errorf("invalid json config (%s): %w", path, err)
		}
		return config, nil
	default:
		return nil, fmt.Errorf("unhandled filetype for plugin file (%s)", filepath.Ext(path))
	}
}

// Get the language set in the config
func GetConfiguredLanguage() language.Tag {
	globalConfig := GlobalConfig()

	switch globalConfig.UserPreferences.Language {
	case config_proto.Language_English:
		return language.English
	case config_proto.Language_Dutch:
		return language.Dutch
	case config_proto.Language_French:
		return language.French
	case config_proto.Language_Spanish:
		return language.Spanish
	default:
		return language.English
	}
}
