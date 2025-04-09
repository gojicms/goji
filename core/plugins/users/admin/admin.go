package admin

import (
	_ "embed"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services"
	"github.com/gojicms/goji/core/types"
	"github.com/gojicms/goji/core/utils"
)

//go:embed listing.gohtml
var listingHtml []byte

//go:embed editor.gohtml
var editorHtml []byte

func Register() {
	extend.AddSideMenuItem("Users", "users", 10, "System", "user:view")
	extend.AddSideMenuItem("Groups", "groups", 20, "System", "admin")

	extend.AddAdminPage(extend.AdminPage{
		Permission: "user:view",
		Route:      "users",
		Render: func(flow *httpflow.HttpFlow) ([]byte, error) {
			flow.Append("templateData", "title", "Goji - Users")

			userService := services.GetServiceOfType[services.UserService]("users")

			offset := utils.OrDefault(flow.Request.URL.Query().Get("offset"), "0")
			count := utils.OrDefault(flow.Request.URL.Query().Get("count"), "10")

			userItems, _ := userService.GetAll()
			userCount, _ := userService.Count()

			content, err := server.RenderTemplate(listingHtml, utils.Object{
				"items":     userItems,
				"offset":    offset,
				"count":     count,
				"itemCount": userCount,
			}, server.DefaultRenderOptions)
			if err != nil {
				d := []byte(fmt.Sprintf("<b>%s</b>", err.Error()))
				return d, nil
			}
			return content, nil
		},
	})

	extend.AddAdminPage(extend.AdminPage{
		Permission: "user:add",
		Route:      "users/new",
		Render: func(flow *httpflow.HttpFlow) ([]byte, error) {
			flow.Append("templateData", "title", "Goji - Create User")

			userService := services.GetServiceOfType[services.UserService]("users")
			groupService := services.GetServiceOfType[services.GroupService]("groups")
			allGroups, err := groupService.GetAll()
			if err != nil {
				d := []byte(fmt.Sprintf("<b>%s</b>", err.Error()))
				return d, nil
			}

			result := utils.Object{
				"status":  nil,
				"message": nil,
			}

			if flow.Request.Method == "POST" {
				result["status"] = "success"
				result["message"] = "Document created."

				displayName := flow.PostFormValue("display_name")
				userName := flow.PostFormValue("user_name")
				email := flow.PostFormValue("email")
				group := flow.PostFormValue("group")
				password := flow.PostFormValue("password")

				if displayName == "" || userName == "" || password == "" {
					result["status"] = "error"
					result["message"] = "Display name or Username or Password is empty."
					goto render
				}

				var user = types.User{
					Username:    userName,
					DisplayName: displayName,
					Password:    password,
					GroupName:   group,
					Email:       email,
				}

				_, err = userService.Create(&user)
			}
		render:
			content, err := server.RenderTemplate(editorHtml, utils.Object{
				"groups": allGroups,
				"create": true,
				"result": result,
			}, server.DefaultRenderOptions)
			if err != nil {
				d := []byte(fmt.Sprintf("<b>%s</b>", err.Error()))
				return d, nil
			}
			return content, nil
		},
	})

	extend.AddAdminPage(extend.AdminPage{
		Permission: "user:view",
		Route:      "users/{id}",
		Render: func(flow *httpflow.HttpFlow) ([]byte, error) {
			flow.Append("templateData", "title", "Goji - Edit User")

			id := flow.GetKvp("admin_meta", "id")
			idInt, _ := strconv.Atoi(id)

			result := utils.Object{
				"status":  nil,
				"message": nil,
			}

			userService := services.GetServiceOfType[services.UserService]("users")
			user, err := userService.GetByID(uint(idInt))
			if err != nil {
				d := []byte(fmt.Sprintf("<b>%s</b>", err.Error()))
				return d, nil
			}

			groupService := services.GetServiceOfType[services.GroupService]("groups")
			allGroups, err := groupService.GetAll()
			if err != nil {
				d := []byte(fmt.Sprintf("<b>%s</b>", err.Error()))
				return d, nil
			}

			if flow.Request.Method == "POST" {
				action := flow.PostFormValue("action")

				if action == "delete" {
					err := userService.Delete(user)
					if err != nil {
						result["status"] = "error"
						result["message"] = "Failed to delete user: " + err.Error()
						goto render
					}
					flow.Redirect("/admin/users", http.StatusFound)
				}
				if action == "save" {
					result["status"] = "success"
					result["message"] = "User updated successfully"

					displayName := flow.PostFormValue("display_name")
					group := flow.PostFormValue("group")
					password := flow.PostFormValue("password")
					email := flow.PostFormValue("email")

					groupObj, err := groupService.GetByName(group)
					if err != nil {
						result["status"] = "error"
						result["message"] = "Failed to update user: " + err.Error()
						goto render
					}

					user.Group = groupObj
					user.GroupName = group
					user.DisplayName = displayName
					user.Email = email

					if password != "" {
						user.Password = password
					}

					err = userService.Update(user)
					if err != nil {
						result["status"] = "error"
						result["message"] = "Failed to update user: " + err.Error()
					}
				}
			}

		render:
			content, err := server.RenderTemplate(editorHtml, utils.Object{
				"user":   user,
				"groups": allGroups,
				"result": result,
			}, server.DefaultRenderOptions)
			if err != nil {
				d := []byte(fmt.Sprintf("<b>%s</b>", err.Error()))
				return d, nil
			}
			return content, nil
		},
	})
}
