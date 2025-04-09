package health

import (
	"github.com/gojicms/goji/core/extend"
)

var Plugin = extend.PluginDef{
	Name:         "health",
	FriendlyName: "Health",
	Description:  "Health monitoring and reporting",
	Internal:     true,
	Resources:    []extend.ResourceDef{},
	OnInit: func() error {
		return nil
	},
}
