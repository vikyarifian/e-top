package handlers

import (
	"etop/models"
	"etop/templates/components/ui"
	"etop/templates/features"
	"etop/utils"
	"net/http"
)

func GetWorkspaces() []models.Workspace {
	ws := []models.Workspace{}
	ws = append(ws, models.Workspace{ID: "1", Name: "IT Dept"})
	ws = append(ws, models.Workspace{ID: "2", Name: "Legal"})
	ws = append(ws, models.Workspace{ID: "3", Name: "Finance"})

	return ws
}

func WorkspaceSwitcher(w http.ResponseWriter, r *http.Request) error {
	ws := GetWorkspaces()
	list := []ui.SelectOption{}
	for _, w := range ws {
		list = append(list, ui.SelectOption{Label: w.Name, Value: w.ID, Color: utils.LetterToColor(w.Name[0:1])})
	}

	return utils.Render(w, r, features.WorkspaceSwitcher(list))
}
