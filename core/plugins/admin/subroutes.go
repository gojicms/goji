package admin

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services"
	"github.com/gojicms/goji/core/types"
	"github.com/gojicms/goji/core/utils"
)

//go:embed 403.gohtml
var forbiddenTemplate []byte

//go:embed 404.gohtml
var notFoundTemplate []byte

func sideNavToScopedObject(user types.User, path string, sidenav []*extend.SideMenuItem) []utils.Object {
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
	path := r.URL.Path

	fileService := services.GetServiceOfType[services.FileService]("files")

	// Get the current user
	user := flow.Get("user").(*types.User)
	flow.Append("templateData", "sideNav", sideNavToScopedObject(*user, path, extend.GetSideMenuItems()))

	// Check for admin page routes
	if route, extras := extend.GetAdminPageByRoute(path); route != nil {
		if user.HasPermission(route.Permission) {
			// Add any extra data from the route
			for k, v := range extras {
				flow.Append("admin_meta", k, v)
			}

			// Try to render the page content
			content, err := route.Render(flow)
			if err != nil {
				renderError(flow, http.StatusInternalServerError, "Failed to render page: "+err.Error())
				return
			}

			// Add the content to the template data
			flow.Append("templateData", "contents", string(content))
		} else {
			content, err := fileService.ExecuteTemplate(forbiddenTemplate, flow)
			if err != nil {
				renderError(flow, http.StatusInternalServerError, "Failed to render forbidden page: "+err.Error())
				return
			}

			flow.Append("templateData", "contents", string(content))

			return
		}
	} else {
		content, err := fileService.ExecuteTemplate(notFoundTemplate, flow)
		if err != nil {
			// If we can't execute the 404 template, fall back to plain text
			flow.WriteHeaders(http.StatusNotFound)
			flow.SetHeader("Content-Type", "text/plain")
			_, _ = flow.Write([]byte("The requested admin page could not be found"))
			return
		}

		// Add the content to the template data
		flow.Append("templateData", "contents", string(content))
	}

	// Render the editor template with the content from templateData
	if err := fileService.RenderTemplateFromPath("admin/editor.html", flow); err != nil {
		renderError(flow, http.StatusInternalServerError, "Failed to render editor: "+err.Error())
		return
	}
}
