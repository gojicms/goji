package admin

import (
	_ "embed"
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

//go:embed "dashboard.gohtml"
var dashboardTemplate []byte

//////////////////////////////////
// Private Methods - Handlers   //
//////////////////////////////////

func resHandler(flow *httpflow.HttpFlow) {
	r := flow.Request
	w := flow.Writer

	path := r.URL.Path

	data := utils.Object{
		"isAuthenticated": sessions.IsAuthenticated(r),
	}

	res, err := server.RenderFile(path, server.RenderOptions{
		Data: data,
	})
	if err != nil {
		res := server.RenderErrorPage(err.HttpCode, err.Message, server.RenderOptions{
			Data: data,
		})
		w.WriteHeader(res.HttpCode)
		_, _ = w.Write(res.Body)
		return
	}

	w.Header().Set("Content-Type", res.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(200)
	_, _ = w.Write(res.Body)
}

var renderLoginPage = func(flow *httpflow.HttpFlow) {
	templateData := flow.Get("templateData")
	if templateData == nil {
		templateData = utils.Object{}
	}

	res, err := server.RenderFile("admin/login.html", server.RenderOptions{
		TemplateRoot: "admin/!partials",
		Data:         templateData.(utils.Object),
	})

	if err != nil {
		res := server.RenderErrorPage(err.HttpCode, err.Message, server.RenderOptions{})
		flow.WriteHeaders(res.HttpCode)
		_, _ = flow.Write(res.Body)
		return
	}

	flow.SetHeader("Content-Type", res.ContentType)
	flow.WriteHeaders(200)
	_, _ = flow.Write(res.Body)
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

	loginError := "Invalid username or password"

	if username == "" || password == "" {
		loginError = "Username or password is empty"
	}

	user, err := users.ValidateLogin(username, password)
	if err != nil {
		flow.Append("templateData", "error", loginError)
		renderLoginPage(flow)
		return
	}
	log.Debug("Admin", "User Found", user)

	if user.HasPermission("admin") == false {
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

var rootResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodGet, "/admin/?"),
	Handler:       rootHandler,
}

//////////////////////////////////
// Resource Definitions         //
//////////////////////////////////

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
// Service Definition           //
//////////////////////////////////

var Service = extend.ServiceDef{
	Name:         "administration",
	FriendlyName: "Administration Panel Service",
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
		extend.AddSideMenuItem("System", "#", 500, "", "admin")
		extend.AddSideMenuItem("Logout", "logout", 1000, "System", "")

		extend.AddAdminPage(extend.AdminPage{
			Route: "dashboard",
			Render: func(flow *httpflow.HttpFlow) ([]byte, error) {
				flow.Append("templateData", "title", "Goji - Welcome")
				return server.RenderTemplate(dashboardTemplate, flow.Get("templateData"), server.DefaultRenderOptions)
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
