package extend

import (
	"strings"

	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
)

type AdminPage struct {
	Route      string
	Render     func(flow *httpflow.HttpFlow) ([]byte, error)
	Permission string
}

var AdminPages []AdminPage

func AddAdminPage(AdminPage AdminPage) {
	AdminPages = append(AdminPages, AdminPage)
	log.Debug("AdminPages", "Adding AdminPage: %s", AdminPage.Route)
}

func GetAdminPages() []AdminPage {
	return AdminPages
}

func GetAdminPageByRoute(route string) (*AdminPage, utils.Object) {
	for _, AdminPage := range AdminPages {
		if matches, extras := matchesPattern(route, "/admin/"+AdminPage.Route); matches == true {
			log.Debug("AdminPages", "Found AdminPage: %s", AdminPage.Route)
			return &AdminPage, extras
		}
	}
	log.Debug("AdminPages", "No AdminPage found for route %s", route)
	return nil, nil
}

func matchesPattern(path, pattern string) (bool, utils.Object) {
	pathParts := strings.Split(path, "/")
	patternParts := strings.Split(pattern, "/")

	if len(pathParts) != len(patternParts) {
		return false, nil
	}

	params := utils.Object{}

	for i := range pathParts {
		// Support named path vars
		if strings.HasPrefix(patternParts[i], "{") && strings.HasSuffix(patternParts[i], "}") {
			paramName := patternParts[i][1 : len(patternParts[i])-1]
			params[paramName] = pathParts[i]
			continue
		}
		if pathParts[i] != patternParts[i] {
			return false, nil
		}
	}

	return true, params
}
