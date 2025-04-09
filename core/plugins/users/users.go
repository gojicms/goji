package users

import (
	"encoding/base64"
	"net/http"

	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/plugins/sessions"
	"github.com/gojicms/goji/core/plugins/users/admin"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services"
	"github.com/gojicms/goji/core/types"
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
		userService := services.GetServiceOfType[services.UserService]("users")

		data := make(map[string]interface{})
		err := flow.DecodeJSONBody(&data)
		if err != nil {
			flow.WriteErrorJson(http.StatusInternalServerError, "Failed to decode JSON body: %s", err.Error())
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

		user, err := userService.ValidateLogin(data["username"].(string), data["password"].(string))

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

var Plugin = extend.PluginDef{
	Name:         "authentication",
	FriendlyName: "Authentication",
	Resources: []extend.ResourceDef{
		loginResource,
		logoutResource,
	},
	OnInit: func() error {
		admin.Register()

		database.AutoMigrate(&types.User{})
		database.AutoMigrate(&types.Group{})

		// Register our service providers
		services.RegisterService(&UserProvider{})
		services.RegisterService(&GroupProvider{})

		// Get our service providers
		userService := services.GetServiceOfType[services.UserService]("users")
		groupService := services.GetServiceOfType[services.GroupService]("groups")

		// Create or update the default groups
		defaultGroups := []struct {
			name        string
			permissions utils.CSV
		}{
			{
				name: "administrator",
				permissions: utils.CSV{
					"admin",
					"document:view", "document:add", "document:edit", "document:delete",
					"media:view", "media:add", "media:edit", "media:delete",
					"user:view", "user:add", "user:edit", "user:delete"},
			},
			{
				name: "editor",
				permissions: utils.CSV{
					"admin",
					"document:view", "document:add", "document:edit", "document:delete",
					"media:view", "media:add", "media:edit", "media:delete"},
			},
			{
				name:        "user",
				permissions: utils.CSV{},
			},
		}

		for _, group := range defaultGroups {
			existingGroup, err := groupService.GetByName(group.name)
			if err != nil {
				// Group doesn't exist, create it
				_ = groupService.Create(&types.Group{
					Name:        group.name,
					Permissions: group.permissions,
					Internal:    true,
				})
			} else {
				// Group exists, update it
				existingGroup.Permissions = group.permissions
				existingGroup.Internal = true
				_ = groupService.Update(existingGroup)
			}
		}

		// Ensure at least one user exists!
		if c, _ := userService.Count(); c == 0 {
			group, err := groupService.GetByName("administrator")
			if err != nil {
				log.Fatal(log.RCDatabase, "Auth", "Failed to create default user, administrator group does not exist")
			}

			password := base64.StdEncoding.EncodeToString([]byte(uuid.New().String()))[:12]
			adminUser := types.User{
				Username:    "admin",
				Password:    password,
				DisplayName: "Goji Admin",
				Group:       group,
			}
			_, err = userService.Create(&adminUser)

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
