package health

import (
	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/utils/log"
)

var Service = extend.ServiceDef{
	Name:         "health",
	FriendlyName: "Health Service",
	Resources:    nil,
	Internal:     true,
	OnInit: func() error {
		extend.AddMiddleware(extend.NewMiddleware("*", "*", 0, func(flow *httpflow.HttpFlow) {
			var errors []string
			if b, err := database.IsReadOnly(); b == false || err != nil {
				log.Error("Health/Database", "Database write check failed - ensure that the database is not loaded in read-only mode and that the connection string is for a user with full database access: %s", err)
				errors = append(errors, "Database access is insufficient. See log for more details.")
			}

			if len(errors) > 0 {
				flow.Append("templateData", "site_errors", errors)
			}
		}))
		return nil
	},
}
