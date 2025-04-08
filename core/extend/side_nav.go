package extend

import (
	"sort"

	"github.com/gojicms/goji/core/utils/log"
)

type SideMenuItem struct {
	Title      string          `json:"title"`
	Path       string          `json:"path"`
	Children   []*SideMenuItem `json:"children"`
	Priority   int             `json:"priority"`
	Permission string          `json:"permission"`
}

var rootSideMenu []*SideMenuItem

func AddSideMenuItem(title, path string, priority int, parent, permission string) {
	sideMenuItem := SideMenuItem{
		Title:      title,
		Path:       "/admin/" + path,
		Priority:   priority,
		Permission: permission,
	}

	// For paths that are #, do not make these functional; they exist as child-only menu items.
	if path == "#" {
		sideMenuItem.Path = "#"
	}

	if parent != "" {
		for _, child := range rootSideMenu {
			if child.Title == parent {
				child.Children = append(child.Children, &sideMenuItem)
				sort.Slice(child.Children, func(i, j int) bool {
					return child.Children[i].Priority < child.Children[j].Priority
				})
				return
			}
		}
		log.Fatal(log.RCAdminConfig, "Admin/Extend", "Attempt to add a side menu to a menu item that does not exist: Parent{%s}", parent)
	}

	rootSideMenu = append(rootSideMenu, &sideMenuItem)
	sort.Slice(rootSideMenu, func(i, j int) bool {
		return rootSideMenu[i].Priority < rootSideMenu[j].Priority
	})
}

func GetSideMenuItems() []*SideMenuItem {
	return rootSideMenu
}
