package context

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Acedyn/zorro-core/internal/processor"
	"github.com/Acedyn/zorro-core/internal/plugin"

	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
	context_proto "github.com/Acedyn/zorro-proto/zorroprotos/context"
	plugin_proto "github.com/Acedyn/zorro-proto/zorroprotos/plugin"
)

// Wrapped context with methods attached
type Context struct {
  *context_proto.Context
}

func (context *Context) GetPlugins() []*plugin.Plugin {
  return slices.Map(context.Context.GetPlugins(), func (p *plugin_proto.Plugin) *plugin.Plugin {
    return &plugin.Plugin{Plugin: p}
  })
}

// Gather the environment variables of all the context's plugins
// in the form "key=value".
func (context *Context) Environ(includeCurrent bool) []string {
  environ := map[string]string{}
  // Convert the current environment to a map rather than
  // a "key=value" slice
  if includeCurrent {
    for _, environVariable := range os.Environ() {
      if splittedEnviron := strings.Split(environVariable, "="); len(splittedEnviron) == 2 {
        environ[splittedEnviron[0]] = splittedEnviron[1]
      }
    }
  }

  // Each plugins brings its own set of environment variable
  // modifications
  for _, plugin := range context.GetPlugins() {
    for key, pluginEnviron := range plugin.GetEnv() {
      // Prepend means insert at the beginning of the current value
      for _, valuePrepend := range pluginEnviron.GetPrepend() {
        if current, ok := environ[key]; ok {
          current = strings.Trim(current, string(filepath.ListSeparator))
          environ[key] = strings.Join([]string{valuePrepend, current}, string(filepath.ListSeparator))
        } else {
          environ[key] = valuePrepend
        }
      }
      // Append means add at the end of the current value
      for _, valueAppend := range pluginEnviron.GetAppend() {
        if current, ok := environ[key]; ok {
          current = strings.Trim(current, string(filepath.ListSeparator))
          environ[key] = strings.Join([]string{current, valueAppend}, string(filepath.ListSeparator))
        } else {
          environ[key] = valueAppend
        }
      }
      // Set will override the current value
      if pluginEnviron.Set != nil {
        environ[key] = pluginEnviron.GetSet()
      }
    }
  }

  // Reformat the environment variables to the "key=value" slice format
  return slices.Map(maps.Keys(environ), func(el string) string {
    return el + "=" + environ[el]
  })
}

// Flatten list of all the processors present in the selected plugins
func (context *Context) AvailableProcessors() []*processor.Processor {
  availableProcessors := []*processor.Processor{}
  for _, plugin := range context.GetPlugins() {
    for _, processor := range plugin.GetProcessors() {
      availableProcessors = append(availableProcessors, processor)
    }
  }

  return availableProcessors
}
