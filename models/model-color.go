package models

type ColorOption struct {
	Name  string `json:"name"`
	Class string `json:"class"`
}

func GetColorOptions() []ColorOption {
	var colors []ColorOption
	colors = append(colors, ColorOption{Name: "blue", Class: "bg-blue-500 dark:bg-blue-400"})
	colors = append(colors, ColorOption{Name: "red", Class: "bg-red-500 dark:bg-red-400"})
	colors = append(colors, ColorOption{Name: "yellow", Class: "bg-yellow-500 dark:bg-yellow-400"})
	colors = append(colors, ColorOption{Name: "green", Class: "bg-green-500 dark:bg-green-400"})
	colors = append(colors, ColorOption{Name: "purple", Class: "bg-purple-500 dark:bg-purple-400"})
	colors = append(colors, ColorOption{Name: "orange", Class: "bg-orange-500 dark:bg-orange-400"})
	colors = append(colors, ColorOption{Name: "teal", Class: "bg-teal-500 dark:bg-teal-400"})

	return colors
}
