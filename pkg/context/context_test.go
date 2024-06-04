package context

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Acedyn/zorro-core/internal/plugin"

	filesystem_proto "github.com/Acedyn/zorro-proto/gen/protos/filesystem"
	plugin_proto "github.com/Acedyn/zorro-proto/gen/protos/plugin"
	"github.com/life4/genesis/slices"
)

// Mocked context
var testContext = Context{
	plugins: []*plugin.Plugin{
		{
			Plugin: &plugin_proto.Plugin{
				Name: "plugin-a",
				Repository: &filesystem_proto.FileSystemSettings{
					FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
				},
				Env: map[string]*plugin_proto.PluginEnv{
					"FOO": {
						Append: []string{
							"/plugin-a/a",
							"/plugin-a/b",
						},
					},
					"BAR": {
						Prepend: []string{
							"/plugin-a/a",
							"/plugin-a/b",
						},
					},
				},
			},
		},
		{
			Plugin: &plugin_proto.Plugin{
				Name: "plugin-b",
				Repository: &filesystem_proto.FileSystemSettings{
					FileSystemSettings: &filesystem_proto.FileSystemSettings_Memory{},
				},
				Env: map[string]*plugin_proto.PluginEnv{
					"FOO": {
						Prepend: []string{
							"/plugin-b/a",
							"/plugin-b/b",
						},
					},
					"BAZ": {
						Set: &[]string{"plugin-b"}[0],
					},
				},
			},
		},
	},
}

// Test the resolution of a environment of a context
func TestEnviron(t *testing.T) {
	resolvedEnviron := testContext.Environ(false)
	expectedEnviron := map[string]string{
		"FOO": strings.Join([]string{
			strings.ReplaceAll(filepath.Join("/plugin-b/b"), string(filepath.Separator), "/"),
			strings.ReplaceAll(filepath.Join("/plugin-b/a"), string(filepath.Separator), "/"),
			strings.ReplaceAll(filepath.Join("/plugin-a/a"), string(filepath.Separator), "/"),
			strings.ReplaceAll(filepath.Join("/plugin-a/b"), string(filepath.Separator), "/"),
		}, string(filepath.ListSeparator)),
		"BAR": strings.Join([]string{
			strings.ReplaceAll(filepath.Join("/plugin-a/b"), string(filepath.Separator), "/"),
			strings.ReplaceAll(filepath.Join("/plugin-a/a"), string(filepath.Separator), "/"),
		}, string(filepath.ListSeparator)),
		"BAZ": "plugin-b",
	}

	for key, environ := range expectedEnviron {
		if !slices.Contains(resolvedEnviron, key+"="+environ) {
			t.Errorf("No resolved environ matched the expected environ %q", environ)
		}
	}
}
