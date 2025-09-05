package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/templates/components/ui"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
	"etop/utils"

	"gorm.io/gorm"
)

func HandleWorkspaceSwitcher(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	var ws []models.Workspace
	db.PgSql.Where("id in (SELECT workspace_id FROM workspace_members WHERE user_id=?)", user.ID).Order("no").Find(&ws)
	if len(ws) == 0 {
		ws = append(ws, models.Workspace{ID: " ", Name: " "})
	}

	list := []ui.SelectOption{}
	for _, w := range ws {
		list = append(list, ui.SelectOption{Label: w.Name, Value: w.ID, Color: w.Color})
	}

	return utils.Render(w, r, features.WorkspaceSwitcher(list))
}

func HandleWorkspaces(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var ws models.Workspace
			if err := db.PgSql.Where("id=?", id).
				Preload("Members", func(db *gorm.DB) *gorm.DB {
					return db.Preload("User")
				}).Preload("Projects", func(db *gorm.DB) *gorm.DB {
				return db.Preload("Members")
			}).Order("no").First(&ws).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return layouts.Layout("404 Not Found", user, pages.NotFound()).Render(r.Context(), w)
			}
			authorized := false
			for _, member := range ws.Members {
				if member.UserID == user.ID {
					authorized = true
					break
				}
			}
			if !authorized {
				return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
			}
			return layouts.Layout("Workspace", user, features.Workspace(ws)).Render(r.Context(), w)
		}

		var ws []models.Workspace
		db.PgSql.Where("id in (SELECT workspace_id FROM workspace_members WHERE user_id=?)", user.ID).
			Preload("Members", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Order("no").Find(&ws)
		return layouts.Layout("Workspaces", user, features.Workspaces(ws)).Render(r.Context(), w)
	case http.MethodPost:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var ws models.Workspace
			if err := db.PgSql.Where("id=?", id).
				Preload("Members", func(db *gorm.DB) *gorm.DB {
					return db.Preload("User")
				}).Preload("Projects", func(db *gorm.DB) *gorm.DB {
				return db.Preload("Members")
			}).Order("no").First(&ws).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				// return ui.Toast("workspace-error", "warning", "", "Workspace not found!", "", nil).Render(r.Context(), w)
				return pages.NotFound().Render(r.Context(), w)
			}
			authorized := false
			for _, member := range ws.Members {
				if member.UserID == user.ID {
					authorized = true
					break
				}
			}
			if !authorized {
				return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
			}
			return features.Workspace(ws).Render(r.Context(), w)
		}

		var ws []models.Workspace
		db.PgSql.Where("id in (SELECT workspace_id FROM workspace_members WHERE user_id=?)", user.ID).Preload("Members").Order("no").Find(&ws)
		return features.Workspaces(ws).Render(r.Context(), w)
	default:
		return nil
	}
}

func HandleWorkspace(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		// var workspace models.Workspace
		// id := r.URL.Query().Get("id")
		// user, _ := auth.GetAuth(w, r)

		// if err := db.PgSql.Where("id=?", id).First(&workspace).Error; err != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	return ui.Toast("workspace-error", "warning", "", "Workspace not found!", "", nil).Render(r.Context(), w)
		// }

		// var member models.WorkspaceMember
		// if err := db.PgSql.Where("id=? AND user_id=?", workspace.ID, user.ID).First(&member).Error; err != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	return ui.Toast("workspace-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		// }

		return nil
	case http.MethodPost:
		var workspace models.Workspace
		r.ParseForm()

		workspace.Name = r.FormValue("name")
		workspace.Description = r.FormValue("description")
		workspace.Color = r.FormValue("color")

		if len(strings.Trim(workspace.Name, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "Name is required!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(workspace.Color, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "Color is required!", "", nil).Render(r.Context(), w)
		}
		no := 0
		user, _ := auth.GetAuth(w, r)
		t := time.Now()
		db.PgSql.Table("workspaces").Select("max(no)").Row().Scan(&no)
		workspace.ID = utils.GenerateHash(strconv.Itoa(no + 1))
		workspace.CreatedAt = &t
		workspace.CreatedBy = user.ID
		workspace.UpdatedAt = &t
		workspace.UpdatedBy = user.ID

		if err := db.PgSql.Create(&workspace).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("workspace-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		members := []models.WorkspaceMember{}
		member := models.WorkspaceMember{}
		no = 0
		db.PgSql.Table("workspace_members").Select("max(no)").Row().Scan(&no)
		member.ID = utils.GenerateHash(strconv.Itoa(no + 1))
		member.WorkspaceID = workspace.ID
		member.UserID = user.ID
		member.Role = "ADMIN"
		member.CreatedAt = &t
		member.CreatedBy = user.ID
		member.UpdatedAt = &t
		member.UpdatedBy = user.ID

		members = append(members, member)

		var membersPayload []string
		if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
			for i, m := range membersPayload {
				members = append(members, models.WorkspaceMember{
					ID:          utils.GenerateHash(strconv.Itoa(no + 2 + i)),
					WorkspaceID: workspace.ID,
					Role:        "MEMBER",
					UserID:      m,
					CreatedAt:   &t,
					CreatedBy:   user.ID,
					UpdatedAt:   &t,
					UpdatedBy:   user.ID,
				})
			}
		}

		db.PgSql.Create(&members)
		return ui.Toast("workspace-success", "success", "", "Workspace created successfully.", "", nil).Render(r.Context(), w)
	case http.MethodPut:
		var workspace models.Workspace
		r.ParseForm()

		workspace.ID = r.FormValue("id")

		user, _ := auth.GetAuth(w, r)
		t := time.Now()
		if err := db.PgSql.Where("id=?", workspace.ID).First(&workspace).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "Workspace is not found!", "", nil).Render(r.Context(), w)
		}

		var member models.WorkspaceMember
		if err := db.PgSql.Where("id=? AND user_id=?", workspace.ID, user.ID).First(&member).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

		workspace.Name = r.FormValue("name")
		workspace.Description = r.FormValue("description")
		workspace.Color = r.FormValue("color")

		if len(strings.Trim(workspace.Name, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "Name is required!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(workspace.Color, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "Color is required!", "", nil).Render(r.Context(), w)
		}

		if strings.ToUpper(strings.Trim(member.Role, " ")) != "ADMIN" {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

		workspace.UpdatedAt = &t
		workspace.UpdatedBy = user.ID

		if err := db.PgSql.Save(&workspace).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("workspace-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}
		return ui.Toast("workspace-success", "success", "", "Workspace updated successfully.", "", nil).Render(r.Context(), w)
	default:
		return nil
	}
}
