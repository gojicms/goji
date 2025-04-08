package auth

import (
	"encoding/base64"
	"net/http"

	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services/auth/admin"
	"github.com/gojicms/goji/core/services/auth/groups"
	"github.com/gojicms/goji/core/services/auth/users"
	"github.com/gojicms/goji/core/services/sessions"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
	"github.com/google/uuid"
)

//////////////////////////////////
// Resource Definitions         //
//////////////////////////////////

var loginResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator("GET", "/admin/login"),
	Description:   "Logs in to the application",
	Handler: func(flow *httpflow.HttpFlow) {
		data := make(map[string]interface{})
		err := flow.DecodeJSONBody(&data)
		if err != nil {
			flow.WriteErrorJson(http.StatusInternalServerError, err.Error())
		}

		username := data["username"].(string)
		password := data["password"].(string)
		csrf := data["_CSRF"].(string)

		if csrf == "" {
			flow.WriteErrorJson(http.StatusBadRequest, "csrf is missing")
		}

		if username == "" || password == "" {
			flow.WriteErrorJson(http.StatusBadRequest, "username or password is empty")
		}

		user, err := users.ValidateLogin(data["username"].(string), data["password"].(string))

		if err != nil {
			flow.WriteErrorJson(http.StatusForbidden, "username or password is invalid")
		}

		session, _ := sessions.CreateSession(flow, csrf, user.ID)
		flow.WriteJson(utils.Object{
			"session": session.SessionId,
		})
	},
}

var logoutResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator("GET", "/admin/logout"),
	Description:   "Logout from the application",
	Handler: func(flow *httpflow.HttpFlow) {
		sessions.EndSession(flow)
		flow.WriteJson(utils.Object{"success": true})
	},
}

//////////////////////////////////
// Service Definition           //
//////////////////////////////////

var Service = extend.ServiceDef{
	Name:         "authentication",
	FriendlyName: "Authentication",
	Resources: []extend.ResourceDef{
		loginResource,
		logoutResource,
	},
	OnInit: func() error {
		admin.Register()

		database.AutoMigrate(&users.User{})
		database.AutoMigrate(&groups.Group{})

		// Ensure the default groups exist
		if c, _ := groups.Count(); c == 0 {
			_ = groups.Create(&groups.Group{
				Name: "administrator",
				Permissions: utils.CSV{
					"admin",
					"user:view", "user:edit", "user:delete", "user:add",
					"document:view", "document:add", "document:edit", "document:delete"},
			})
			_ = groups.Create(&groups.Group{
				Name:        "editor",
				Permissions: utils.CSV{"admin", "document:view", "document:add", "document:edit", "document:delete"},
			})
			_ = groups.Create(&groups.Group{
				Name:        "user",
				Permissions: utils.CSV{},
			})
			if c, _ = groups.Count(); c == 0 {
				log.Fatal(log.RCDatabase, "Auth", "Failed to create default user groups")
			}
		}

		// Ensure at least one user exists!
		if c, _ := users.Count(); c == 0 {
			group, err := groups.GetByName("administrator")
			if err != nil {
				log.Fatal(log.RCDatabase, "Auth", "Failed to create default user, administrator group does not exist")
			}

			password := base64.StdEncoding.EncodeToString([]byte(uuid.New().String()))[:12]
			adminUser := users.User{
				Username:    "admin",
				Password:    password,
				DisplayName: "Goji Admin",
				Group:       group,
			}
			_, err = users.Create(&adminUser)

			if err != nil {
				log.Fatal(log.RCDatabase, "Auth", "Could not create admin user: %s", err)
			}

			// Notify the user of this
			log.Warn("Users", "===================== IMPORTANT =====================")
			log.Warn("Users", "No users exist; an admin user has been created.")
			log.Warn("Users", "The password for this user is: %s", password)
			log.Warn("Users", "Change this IMMEDIATELY!")
			log.Warn("Users", "===================== IMPORTANT =====================")
		}

		return nil
	},
}
