package config

import (
	"time"

	"github.com/gojicms/goji/core/utils/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CmsConfig struct {
	Version    string
	Configured bool
}

type ApplicationConfig struct {
	Host                  string
	Port                  string
	RootUrl               string
	ApiRootUrl            string
	Debug                 bool
	TemplateFileSizeLimit int64
	LogLevel              log.LogLevel
	Database              DatabaseConfig
	Auth                  AuthConfig
	Pepper                string `json:"-"` // Don't provide the pepper EVEN IF DEBUG IS ENABLED!
}

type DatabaseConfig struct {
	Connector func() gorm.Dialector `json:"-"`
	Config    gorm.Config           `json:"-"`
}

type AuthConfig struct {
	// CookieId The name of the cookie to use to store the user's session
	CookieId string
	// CSRFId The name of the cookie to use to store the CSRF value
	CSRFId string
	// CookieLifetime How long the cookie should last
	CookieLifetime time.Duration
	// RefreshLifetime How long the cookie should last before a new cookie is provided.
	RefreshLifetime time.Duration
}

type Config struct {
	Application ApplicationConfig
	Cms         CmsConfig
}

var ActiveConfig Config = Config{
	Cms: CmsConfig{
		Version: "v0.1",
	},
	Application: ApplicationConfig{
		Host:                  "",
		Port:                  "8080",
		RootUrl:               "",
		ApiRootUrl:            "",
		Debug:                 false,
		TemplateFileSizeLimit: 10,
		LogLevel:              log.LogWarn | log.LogError | log.LogInfo,
		Pepper:                "pepper",
		Auth: AuthConfig{
			CookieId:        "Goji_Auth",
			CSRFId:          "Goji_CSRF",
			CookieLifetime:  time.Hour,
			RefreshLifetime: time.Minute * 45,
		},
		Database: DatabaseConfig{
			func() gorm.Dialector { return sqlite.Open("file:mydatabase.db?cache=shared&mode=rwc") },
			gorm.Config{},
		},
	},
}
