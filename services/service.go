package services

import (
	"etop/models"
)

func GetMenu() []models.Menu {
	menus := []models.Menu{}
	menus = append(menus, models.Menu{Label: "Dashboard", Href: "/dashboard", Icon: "home"})
	menus = append(menus, models.Menu{Label: "Department", Href: "/department", Icon: "network"})
	menus = append(menus, models.Menu{Label: "Workspaces", Href: "/workspaces", Icon: "briefcase-business"})
	// menus = append(menus, models.Menu{Label: "Workspaces", Href: "/workspaces", Icon: "monitor-check"})
	menus = append(menus, models.Menu{Label: "My Tasks", Href: "/my-tasks", Icon: "layout-list"})
	// menus = append(menus, models.Menu{Label: "Members", Href: "/members", Icon: "users"})
	menus = append(menus, models.Menu{Label: "Achieved", Href: "/achieve", Icon: "badge-check"})
	menus = append(menus, models.Menu{Label: "Settings", Href: "/settings", Icon: "settings"})

	return menus
}

func GetColorOptions() []models.ColorOption {
	var colors []models.ColorOption
	colors = append(colors, models.ColorOption{Name: "blue", Class: "bg-blue-500 dark:bg-blue-400"})
	colors = append(colors, models.ColorOption{Name: "red", Class: "bg-red-500 dark:bg-red-400"})
	colors = append(colors, models.ColorOption{Name: "yellow", Class: "bg-yellow-500 dark:bg-yellow-400"})
	colors = append(colors, models.ColorOption{Name: "green", Class: "bg-green-500 dark:bg-green-400"})
	colors = append(colors, models.ColorOption{Name: "purple", Class: "bg-purple-500 dark:bg-purple-400"})
	colors = append(colors, models.ColorOption{Name: "orange", Class: "bg-orange-500 dark:bg-orange-400"})
	colors = append(colors, models.ColorOption{Name: "teal", Class: "bg-teal-500 dark:bg-teal-400"})

	return colors
}
