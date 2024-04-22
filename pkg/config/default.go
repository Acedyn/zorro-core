package config

import (
	config_proto "github.com/Acedyn/zorro-proto/gen/protos/config"
	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
)

var DEFAULT = &config_proto.Config{
	UserPreferences: &config_proto.UserConfig{
		Language: config_proto.Language_English,
	},
	PluginConfig: &config_proto.PluginConfig{
		DefaultRequire:     []string{},
		Repositories:       []*filesystem_proto.FileSystemSettings{},
		SearchMaximumDepht: 3,
	},
	NetworkConfig: &config_proto.NetworkConfig{
		GRPCPort: 8686,
		GRPCHost: "127.0.0.1",
	},
}
