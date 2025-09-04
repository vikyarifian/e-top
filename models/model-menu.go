package models

type Menu struct {
	Label string
	Href  string
	Icon  string
}

func GetMenu() []Menu {
	menus := []Menu{}
	menus = append(menus, Menu{Label: "Dashboard", Href: "/dashboard", Icon: "home"})
	menus = append(menus, Menu{Label: "Workspaces", Href: "/workspaces", Icon: "briefcase-business"})
	// menus = append(menus, Menu{Label: "Workspaces", Href: "/workspaces", Icon: "monitor-check"})
	menus = append(menus, Menu{Label: "My Tasks", Href: "/tasks", Icon: "layout-list"})
	menus = append(menus, Menu{Label: "Members", Href: "/members", Icon: "users"})
	menus = append(menus, Menu{Label: "Achieved", Href: "/achieve", Icon: "badge-check"})
	menus = append(menus, Menu{Label: "Settings", Href: "/settings", Icon: "settings"})

	return menus
}
