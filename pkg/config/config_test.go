package config

import (
	"testing"

	_ "github.com/Acedyn/zorro-core/testdata"
	"github.com/life4/genesis/slices"

	config_proto "github.com/Acedyn/zorro-proto/gen/protos/config"
	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
)

type ConfigTest struct {
	Path                 string
	ExpectedLanguage     language.Tag
	ExpectedPluginConfig config_proto.PluginConfig
}

var configTests = []*ConfigTest{
	{
		Path:             "configs/foo.json",
		ExpectedLanguage: language.Spanish,
		ExpectedPluginConfig: config_proto.PluginConfig{
			SearchMaximumDepht: 3,
		},
	},
	{
		Path:             "configs/bar.json",
		ExpectedLanguage: language.French,
		ExpectedPluginConfig: config_proto.PluginConfig{
			SearchMaximumDepht: 2,
			DefaultRequire:     []string{"toto", "tata"},
			Repositories: []*filesystem_proto.FileSystemSettings{
				{
					FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{
						Memory: &filesystem_proto.MemoryFsSettings{
							Name: "in-memory",
						},
					},
				},
			},
		},
	},
}

// Test loading a json config from an in memory file system
func TestLoadJson(t *testing.T) {
	for _, configTest := range configTests {
		loadedConfig, err := LoadConfig(configTest.Path, &filesystem_proto.FileSystemSettings{
			FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
		})

		if err != nil {
			t.Errorf("An error occured while loadding a config from the in memory file system:\n\t%s", err)
			return
		}

		if GetConfiguredLanguage() != configTest.ExpectedLanguage {
			t.Errorf(
				"Unexpected language resolved from config %s (found: %s, expected: %s)",
				configTest.Path,
				GetConfiguredLanguage(),
				configTest.ExpectedLanguage,
			)
			return
		}

		pluginConfigProto, err := proto.Marshal(loadedConfig.PluginConfig)
		if err != nil {
			t.Errorf("Invalid plugin config loaded: %s", err)
			return
		}
		expectedPluginConfigProto, err := proto.Marshal(&configTest.ExpectedPluginConfig)
		if err != nil {
			t.Errorf("Invalid plugin config expected: %s", err)
			return
		}
		if !slices.Equal(pluginConfigProto, expectedPluginConfigProto) {
			t.Errorf(
				"Repositories entry mismatch from config %s (found %s, expected %s)",
				configTest.Path,
				loadedConfig.PluginConfig,
				&configTest.ExpectedPluginConfig,
			)
		}
	}
}
