package context

import (
	"path/filepath"
	"strings"
	"testing"

	context_proto "github.com/Acedyn/zorro-proto/zorroprotos/context"
	plugin_proto "github.com/Acedyn/zorro-proto/zorroprotos/plugin"
	"github.com/life4/genesis/slices"
)

// Mocked context
var testContext = Context{
	Context: &context_proto.Context{
		Plugins: []*plugin_proto.Plugin{
			{
				Name: "plugin-a",
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
			{
				Name: "plugin-b",
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
			"/plugin-b/b",
			"/plugin-b/a",
			"/plugin-a/a",
			"/plugin-a/b",
		}, string(filepath.ListSeparator)),
    "BAR": strings.Join([]string{
			"/plugin-a/b",
			"/plugin-a/a",
		}, string(filepath.ListSeparator)),
    "BAZ": "plugin-b",
	}

	for key, environ := range expectedEnviron {
		if !slices.Contains(resolvedEnviron, key + "=" + environ) {
			t.Errorf("No resolved environ matched the expected environ %q", environ)
		}
	}
}
