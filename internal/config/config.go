package config

import (
	"sync"

	"golang.org/x/text/language"
)

var (
	config *Config
	once   sync.Once
)

// Getter for the config singleton
func AppConfig() *Config {
	once.Do(func() {
		// TODO: Get the config from somewhere on the computer
		config = &Config{
			UserPreferences: &UserConfig{
				Language: Language_English,
			},
		}
	})

	return config
}

// Get the language set in the config
func GetLanguage() language.Tag {
	config = AppConfig()

	switch config.UserPreferences.Language {
	case Language_English:
		return language.English
	case Language_Dutch:
		return language.Dutch
	case Language_French:
		return language.French
	case Language_Spanish:
		return language.Spanish
	default:
		return language.English
	}
}
