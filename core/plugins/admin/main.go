package admin

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/plugins/sessions"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services"
	"github.com/gojicms/goji/core/utils/log"
)

//go:embed "dashboard.gohtml"
var dashboardTemplate []byte

//////////////////////////////////
// Private Methods - Handlers   //
//////////////////////////////////

func resHandler(flow *httpflow.HttpFlow) {
	r := flow.Request

	log.Debug("Admins", "Rendering file: %s", r.URL.Path)

	fileService := services.GetServiceOfType[services.FileService]("files")

	if err := fileService.RenderFile(r.URL.Path, flow); err != nil {
		// RenderFile handles writing the response status and body on error
		log.Error("Admin", "Error rendering file via resHandler: %s - %v", r.URL.Path, err)
		// No need to call renderError here as RenderFile wrote the response
	}
}

var renderLoginPage = func(flow *httpflow.HttpFlow) {
	fileService := services.GetServiceOfType[services.FileService]("files")
	if err := fileService.RenderTemplateFromPath("admin/login.html", flow); err != nil {
		renderError(flow, http.StatusInternalServerError, "File Error: "+err.Error())
		return
	}
}

var loginHandler = func(flow *httpflow.HttpFlow) {
	user := flow.Get("user")
	if user != nil {
		flow.Redirect("/admin/dashboard", http.StatusFound)
	}
	renderLoginPage(flow)
}

var logoutHandler = func(flow *httpflow.HttpFlow) {
	sessions.EndSession(flow)
	flow.Redirect("/admin/login", http.StatusFound)
	return
}

var loginPostHandler = func(flow *httpflow.HttpFlow) {
	username := flow.PostFormValue("username")
	password := flow.PostFormValue("password")
	nonce := flow.PostFormValue("_CSRF")

	userService := services.GetServiceOfType[services.UserService]("users")

	loginError := "Invalid username or password"

	if username == "" || password == "" {
		loginError = "Username or password is empty"
	}

	user, err := userService.ValidateLogin(username, password)
	if err != nil {
		flow.Append("templateData", "error", loginError)
		renderLoginPage(flow)
		return
	}

	if !user.HasPermission("admin") {
		flow.Append("templateData", "error", "You are not an admin and cannot access this page.")
		renderLoginPage(flow)

	}

	_, _ = sessions.CreateSession(flow, nonce, user.ID)
	flow.Redirect("/admin/dashboard", http.StatusFound)
	return
}

var rootHandler = func(flow *httpflow.HttpFlow) {
	if flow.Has("user") {
		flow.Redirect("/admin/dashboard", http.StatusFound)
		return
	}
	flow.Redirect("/admin/login", http.StatusFound)
}

//////////////////////////////////
// Resource Definitions         //
//////////////////////////////////

var rootResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodGet, "/admin/?"),
	Handler:       rootHandler,
}

var publicResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodGet, "/admin/public/.+"),
	Handler:       resHandler,
}

var subRouteResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator("*", "/admin/.+"),
	Handler:       subRouteHandler,
}

var loginResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator("GET", "/admin/login"),
	Handler:       loginHandler,
}

var doLoginResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator("POST", "/admin/login"),
	Handler:       loginPostHandler,
}

var logoutResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator("*", "/admin/logout"),
	Handler:       logoutHandler,
}

//////////////////////////////////
// Plugin Definition           //
//////////////////////////////////

var Plugin = extend.PluginDef{
	Name:         "admin",
	FriendlyName: "Admin",
	Description:  "Administration interface for Goji",
	Internal:     true,
	Resources: []extend.ResourceDef{
		publicResource,
		loginResource,
		doLoginResource,
		logoutResource,
		subRouteResource,
		rootResource,
	},
	OnInit: func() error {
		extend.AddSideMenuItem("Home", "dashboard", 0, "", "")
		extend.AddSideMenuItem("Media", "#", 250, "", "admin")
		extend.AddSideMenuItem("System", "#", 500, "", "admin")
		extend.AddSideMenuItem("Logout", "logout", 1000, "System", "")

		extend.AddAdminPage(extend.AdminPage{
			Route: "dashboard",
			Render: func(flow *httpflow.HttpFlow) ([]byte, error) {
				fileService := services.GetServiceOfType[services.FileService]("files")
				return fileService.ExecuteTemplate(dashboardTemplate, flow)
			},
		})

		extend.AddMiddleware(extend.NewMiddleware("*", "^/admin", 50, func(flow *httpflow.HttpFlow) {
			requestPath := flow.Request.URL.Path

			if strings.HasPrefix(requestPath, "/admin/login") || strings.HasPrefix(requestPath, "/admin/public") {
				return
			}

			if !flow.Has("session") {
				log.Debug("Admin", "Session not found - Directing user to login")
				flow.Redirect("/admin/login", http.StatusFound)
				flow.Terminate()
			}
		}))

		return nil
	},
}
