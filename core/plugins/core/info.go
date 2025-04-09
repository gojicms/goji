package core

import (
	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/utils"
)

var coreInfoResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator("GET", "/api/v1"),
	Handler: func(flow *httpflow.HttpFlow) {
		var services []interface{}

		for _, service := range extend.GetPlugins() {
			services = append(services, service.ToApiJson())
		}

		httpflow.WriteJson(flow, utils.Object{
			"_version": config.ActiveConfig.Cms.Version,
			"services": "",
		})
	},
}
