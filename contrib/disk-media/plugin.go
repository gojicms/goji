package diskmedia

import (
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/services"
)

//////////////////////////////////
// Types                        //
//////////////////////////////////

type diskMediaService struct{}

//////////////////////////////////
// Service Interface Impl       //
//////////////////////////////////

func (p *diskMediaService) Name() string {
	return "disk-media"
}

func (p *diskMediaService) Description() string {
	return "On-device media storage service"
}

func (p *diskMediaService) Priority() int {
	return 10
}

//////////////////////////////////
// Plugin Definition           //
//////////////////////////////////

var Plugin = extend.PluginDef{
	Name:         "disk-media",
	FriendlyName: "Disk Media",
	Description:  "On-device media storage service",
	Internal:     true,
	Resources:    []extend.ResourceDef{},
	OnInit: func() error {
		// Register our media service
		services.RegisterService(&diskMediaService{})

		// Add admin menu item
		extend.AddSideMenuItem("Local", "media/local", -1, "Media", "media:view")

		return nil
	},
}
