package extend

var plugins []*PluginDef

type PluginDef struct {
	Name         string        `json:"name"`
	FriendlyName string        `json:"friendly_name"`
	Description  string        `json:"description"`
	Resources    []ResourceDef `json:"resources"`
	Internal     bool          `json:"internal"`
	OnInit       func() error  `json:"-"`
	HealthCheck  func() error  `json:"-"`
}

func (d PluginDef) ToApiJson() interface{} {
	return map[string]interface{}{
		"name":          d.Name,
		"friendly_name": d.FriendlyName,
		"description":   d.Description,
	}
}

func RegisterPlugin(plugin *PluginDef) {
	plugins = append(plugins, plugin)
}

func GetPlugins() []*PluginDef {
	pluginsCopy := make([]*PluginDef, len(plugins))
	copy(pluginsCopy, plugins)
	return pluginsCopy
}

func GetPlugin(name string) *PluginDef {
	for _, plugin := range plugins {
		if plugin.Name == name {
			return plugin
		}
	}
	return nil
}
