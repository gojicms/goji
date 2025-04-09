package site

import (
	_ "embed"

	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/extend"
)

//go:embed "site.gohtml"
var siteConfigTemplate []byte
var siteConfigCache = make(map[string]string)

//////////////////////////////////
// Types Definitions            //
//////////////////////////////////

type SiteConfig struct {
	Key   string `gorm:"primarykey;size:255"`
	Value string `gorm:"type:text"`
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
// Plugin Definition           //
//////////////////////////////////

var Plugin = extend.PluginDef{
	Name:         "site",
	FriendlyName: "Site",
	Description:  "Site configuration and management",
	Internal:     true,
	Resources:    []extend.ResourceDef{},
	OnInit: func() error {
		return nil
	},
}
