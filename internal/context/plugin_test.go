package context

import (
  "testing"
)

func TestLoadPluginFromFile(t *testing.T) {
  
}

func TestLoadPluginBare(t *testing.T) {
  pluginPath := "/foo/bar@1.2/zorro-plugin.json"
  plugin := LoadPluginBare(pluginPath)

  if plugin.Name != "bar" {
    t.Errorf("Invalid bare plugin's name loaded: %s", plugin.Name)
  }

  if plugin.Version != "1.2" {
    t.Errorf("Invalid bare plugin's version loaded: %s", plugin.Version)
  }
}

func TestPluginInit(t *testing.T) {
  pluginPath := "/foo/bar/zorro-plugin.json"
  plugin := &Plugin{
    Name: "foo_bar",
    Path: &pluginPath,
    Tools: &PluginTools{
      Commands: []string{"./commands", "/foo/commands"},
    },
  }

  plugin.InitFields()

  if *plugin.Label != "Foo Bar" {
    t.Errorf("Invalid plugin label initialized: %s", *plugin.Label)
  }

  if plugin.Tools.Commands[0] != "/foo/bar/commands" {
    t.Errorf("Invalid plugin command paths initialized: %s", plugin.Tools.Commands)
  }

  if plugin.Tools.Commands[1] != "/foo/commands" {
    t.Errorf("Invalid plugin command paths initialized: %s", plugin.Tools.Commands)
  }
}
