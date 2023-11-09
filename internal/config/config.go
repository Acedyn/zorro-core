package config

import (
	"sync"

	"golang.org/x/text/language"
	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
)

var (
	config *config_proto.Config
	once   sync.Once
)

// Getter for the config singleton
func AppConfig() *config_proto.Config {
	once.Do(func() {
		// TODO: Get the config from somewhere on the computer
		config = &config_proto.Config{
			UserPreferences: &config_proto.UserConfig{
				Language: config_proto.Language_English,
			},
		}
	})

	return config
}

// Get the language set in the config
func GetLanguage() language.Tag {
	config = AppConfig()

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
