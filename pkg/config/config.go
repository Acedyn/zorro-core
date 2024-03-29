package config

import (
	"sync"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	"golang.org/x/text/language"
)

var (
	config *config_proto.Config
	once   sync.Once
)

// Getter for the config singleton
func GlobalConfig() *config_proto.Config {
	once.Do(func() {
		// TODO: Get the config from somewhere on the computer
		config = &config_proto.Config{
			UserPreferences: &config_proto.UserConfig{
				Language: config_proto.Language_English,
			},
			PluginConfig: &config_proto.PluginConfig{
				DefaultRequire: []string{},
				Repositories:   []*config_proto.RepositoryConfig{},
			},
			NetworkConfig: &config_proto.NetworkConfig{
				GRPCPort: 8686,
				GRPCHost: "127.0.0.1",
			},
		}
	})

	return config
}

// Get the language set in the config
func GetLanguage() language.Tag {
	config = GlobalConfig()

	switch config.UserPreferences.Language {
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
