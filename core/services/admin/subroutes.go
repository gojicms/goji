package admin

import (
	_ "embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services/auth/users"
	"github.com/gojicms/goji/core/services/sessions"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
)

//go:embed "500.gohtml"
var internalServerError []byte

//go:embed "403.gohtml"
var forbidden []byte

func sideNavToScopedObject(user users.User, path string, sidenav []*extend.SideMenuItem) []utils.Object {
	var menu []utils.Object

	for _, item := range sidenav {
		var children []utils.Object
		var hasActiveChild = false

		if user.HasPermission(item.Permission) {
			for _, child := range item.Children {
				if user.HasPermission(child.Permission) {
					isActive := strings.HasPrefix(path, child.Path)
					hasActiveChild = hasActiveChild || isActive
					children = append(children, utils.Object{
						"title":  child.Title,
						"path":   child.Path,
						"active": isActive,
					})
				}
			}

			// If the top level item has a null path, and has at least one child, have it point to the first child.
			if item.Path == "#" && len(item.Children) > 0 {
				item.Path = item.Children[0].Path
			}

			toAppend := utils.Object{
				"title":    item.Title,
				"path":     item.Path,
				"active":   hasActiveChild || strings.HasPrefix(path, item.Path),
				"children": children,
			}

			menu = append(menu, toAppend)
		}
	}

	return menu
}

func subRouteHandler(flow *httpflow.HttpFlow) {
	r := flow.Request
	w := flow.Writer

	var user *users.User
	user = flow.Get("user").(*users.User)

	flow.Append("templateData", "sideNav", sideNavToScopedObject(*user, flow.Request.URL.Path, extend.GetSideMenuItems()))

	renderServerError := func(serveError *server.HttpServeError) {
		res := server.RenderErrorPage(http.StatusInternalServerError, "Internal Server Error", server.RenderOptions{
			TemplateRoot: "admin/!partials",
			ErrorRoot:    "admin",
			Data: utils.Object{
				"error":           serveError,
				"isAuthenticated": sessions.IsAuthenticated(r),
			},
		})

		w.WriteHeader(res.HttpCode)
		_, _ = w.Write(res.Body)
	}

	if route, extras := extend.GetAdminPageByRoute(r.URL.Path); route != nil {
		var rendered []byte
		var err error

		if route.Permission == "" || user.HasPermission(route.Permission) {
			for k, v := range extras {
				flow.Append("admin_meta", k, v)
			}

			// Use a function that captures panics and returns error info
			func() {
				// Defer panic recovery
				defer func() {
					if r := recover(); r != nil {
						log.Error("Admin", "Failed to render editor: %v", r)
						// Convert panic to error message and set rendered to error HTML
						rendered = internalServerError
						// Set err to nil so we continue to RenderFile
						err = nil
					}
				}()

				// Try to render, this might panic
				rendered, err = route.Render(flow)
			}()
		} else {
			log.Error("Security", "User %s (%d) attempted to access %s", user.DisplayName, user.ID, r.URL.Path)
			rendered = forbidden
		}

		if err != nil {
			renderServerError(nil)
			log.Error("Admin/Subroutes", "An unknown error has occurred", err)
			return
		}

		flow.Append("templateData", "contents", string(rendered))
		flow.Append("templateData", "error", err)

		res, serr := server.RenderFile("admin/editor.html", server.RenderOptions{
			TemplateRoot: "admin/!partials",
			ErrorRoot:    "admin",
			Functions: template.FuncMap{
				"pathClass": func(path string) string {
					return "path"
				},
				"subPathClass": func(path string) string {
					return "subnav"
				},
			},
			Data: flow.Get("templateData").(utils.Object),
		})

		if serr != nil {
			renderServerError(serr)
			return
		}

		w.Header().Set("Content-Type", res.ContentType)
		w.WriteHeader(200)
		_, _ = w.Write(res.Body)
		return
	}

	res := server.RenderErrorPage(http.StatusNotFound, "Not Found", server.RenderOptions{
		TemplateRoot: "admin/!partials",
		ErrorRoot:    "admin",
		Data:         flow.Get("templateData").(utils.Object),
	})
	w.WriteHeader(res.HttpCode)
	_, _ = w.Write(res.Body)
	return
}
