package site

import (
	_ "embed"
	"strings"

	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server"
	"github.com/gojicms/goji/core/server/httpflow"
)

//go:embed "site.gohtml"
var siteConfigTemplate []byte
var siteConfigCache = make(map[string]string)

//////////////////////////////////
// Types Definitions            //
//////////////////////////////////

type SiteConfig struct {
	Key   string `gorm:"primarykey"`
	Value string
}

//////////////////////////////////
// Public Methods Definitions         //
//////////////////////////////////

func SetSiteConfig(key string, value string) {
	db := database.GetDB()
	db.Save(SiteConfig{
		Key:   key,
		Value: value,
	})
	siteConfigCache[key] = value
}

func GetSiteConfig(key string) string {
	if value, ok := siteConfigCache[key]; ok {
		return value
	}

	db := database.GetDB()
	var siteConfig SiteConfig
	db.First(&siteConfig, "key = ?", key)

	siteConfigCache[key] = siteConfig.Value

	return siteConfig.Value
}

func SyncSiteConfigs() {
	db := database.GetDB()
	var siteConfigs []SiteConfig
	db.Find(&siteConfigs)
	siteConfigCache = make(map[string]string)
	for _, siteConfig := range siteConfigs {
		siteConfigCache[siteConfig.Key] = siteConfig.Value
	}
}

//////////////////////////////////
// Service Definition           //
//////////////////////////////////

var Service = extend.ServiceDef{
	Name:         "site_config",
	FriendlyName: "Site Configuration",
	Resources:    []extend.ResourceDef{},
	OnInit: func() error {
		database.AutoMigrate(SiteConfig{})

		SyncSiteConfigs()

		extend.AddSideMenuItem("Site", "site", 0, "", "admin")

		extend.AddAdminPage(extend.AdminPage{
			Route: "site",
			Render: func(flow *httpflow.HttpFlow) ([]byte, error) {
				if flow.Request.Method == "POST" {
					err := flow.Request.ParseForm()
					if err == nil {
						for k, v := range flow.Request.Form {
							SetSiteConfig(k, strings.Join(v, ","))
						}
					}
				}

				data, _ := server.RenderTemplate(siteConfigTemplate, flow.Get("templateData"), server.DefaultRenderOptions)
				return data, nil
			},
			Permission: "",
		})

		extend.AddMiddleware(extend.NewMiddleware("*", "*", 10, func(flow *httpflow.HttpFlow) {
			flow.Append("templateData", "site", siteConfigCache)
		}))

		return nil
	},
}
