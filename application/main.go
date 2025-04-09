package main

import (
	"time"

	diskmedia "github.com/gojicms/goji/contrib/disk-media"
	documents "github.com/gojicms/goji/contrib/documents"
	"github.com/gojicms/goji/core"
	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// This represents a standard installation of Goji
func main() {
	// Prepares the server
	core.PrepareServer(config.ApplicationConfig{
		Host: "0.0.0.0",
		// This determines the route for web content
		RootUrl: "",
		// And this determines the route for APIs
		ApiRootUrl: "/api/v1",
		// If true, certain features will be enabled to help track bugs down
		Debug: true,
		// Sets a maximum size for templates; any templates that exceed this will not be processed
		TemplateFileSizeLimit: 1024 * 1024 * 10,
		// Sets the logs that are enabled
		LogLevel: log.LogError | log.LogWarn | log.LogInfo | log.LogVerbose | log.LogDebug,
		// Configure how authentication works
		Auth: config.AuthConfig{
			// This identifies the name for auth cookies
			CookieLifetime:  time.Hour,
			RefreshLifetime: time.Minute * 45,
		},
		// Configure a basic SQLite Database; Not ideal for production... perhaps?
		Database: config.DatabaseConfig{
			Connector: func() gorm.Dialector {
				return sqlite.Open(utils.GetEnv("DB_DSN", "file:application.db?cache=shared&mode=rwc"))
			},
		},
	})

	// Plugins can be loaded dynamically by placing them in the plugins directory,
	// or you can statically load them here.
	extend.RegisterPlugin(&documents.Plugin)
	extend.RegisterPlugin(&diskmedia.Plugin)

	// Start server
	core.StartServer()
}
