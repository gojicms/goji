package core

import (
	"github.com/gojicms/goji/core/extend"
)

var Plugin = extend.PluginDef{
	Name:         "core",
	FriendlyName: "Core",
	Description:  "Core functionality and APIs",
	Internal:     true,
	Resources: []extend.ResourceDef{
		coreInfoResource,
		publicResResource,
		httpResource,
	},
	OnInit: func() error { return nil },
}
