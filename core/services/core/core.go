package core

import (
	"github.com/gojicms/goji/core/extend"
)

var Service = extend.ServiceDef{
	Name:         "core",
	FriendlyName: "Core",
	Resources: []extend.ResourceDef{
		coreInfoResource,
		publicResResource,
		httpResource,
	},
	OnInit: func() error { return nil },
}
