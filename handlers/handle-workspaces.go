package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/services"
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

func HandleCreateWorkspaceForm(w http.ResponseWriter, r *http.Request) error {
	return features.CreateWorkspaceForm().Render(r.Context(), w)
}

func HandleEditWorkspaceForm(w http.ResponseWriter, r *http.Request) error {
	workspaceID := r.URL.Query().Get("workspace_id")
	workspace := models.Workspace{}
	if err := db.PgSql.Where("id=?", workspaceID).Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User")
	}).First(&workspace).Error; err != nil {
		return ui.Toast("workspace-error", "warning", "", "Workspace not found!", "", nil).Render(r.Context(), w)
	}
	return features.EditWorkspaceForm(workspace).Render(r.Context(), w)
}

func HandleInviteWorkspaceForm(w http.ResponseWriter, r *http.Request) error {
	workspaceID := r.URL.Query().Get("workspace_id")
	workspace := models.Workspace{}
	users := []models.User{}
	invitedUsers := []models.InviteWorkspaceMember{}
	if err := db.PgSql.Where("id=?", workspaceID).Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User")
	}).First(&workspace).Error; err != nil {
		return ui.Toast("workspace-error", "warning", "", "Workspace not found!", "", nil).Render(r.Context(), w)
	}
	db.PgSql.Where("id NOT IN (SELECT user_id FROM workspace_members WHERE workspace_id=?) AND id NOT IN (SELECT user_id FROM invite_workspace_members WHERE workspace_id=? AND status='INVITED')", workspaceID, workspaceID).
		Find(&users)
	db.PgSql.Where("user_id NOT IN (SELECT user_id FROM workspace_members WHERE workspace_id=?) AND status='INVITED' AND workspace_id=?", workspaceID, workspaceID).
		Preload("User").Find(&invitedUsers)

	userRole := []models.UserRole{}
	for _, user := range users {
		var m models.UserRole
		m.ID = user.ID
		m.FullName = user.FullName
		m.Color = user.Color
		userRole = append(userRole, m)
	}
	return features.InviteWorkspaceForm(workspace, userRole, invitedUsers).Render(r.Context(), w)
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
				return db.Where("id IN (SELECT project_id FROM project_members WHERE user_id=?)", user.ID).
					Preload("Members").Preload("Tasks")
			}).Order("no").First(&ws).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return layouts.Layout("404 Not Found", user, pages.NotFound()).Render(r.Context(), w)
			}
			authorized := false
			userRole := ""
			for _, member := range ws.Members {
				if member.UserID == user.ID {
					authorized = true
					userRole = member.Role
					break
				}
			}
			for _, project := range ws.Projects {
				for _, member := range project.Members {
					if member.UserID == user.ID {
						authorized = true
						if userRole == "" {
							userRole = member.Role
						}
						break
					}
				}
			}
			if !authorized {
				return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
			}
			var contributors []models.User
			db.PgSql.Where("id IN (SELECT user_id FROM project_members WHERE role='CONTRIBUTOR' AND project_id IN (SELECT id FROM projects WHERE workspace_id=?))", ws.ID).Find(&contributors)
			return layouts.Layout("Workspace", user, features.Workspace(userRole, ws, contributors)).Render(r.Context(), w)
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
				return db.Where("id IN (SELECT project_id FROM project_members WHERE user_id=?)", user.ID).
					Preload("Members").Preload("Tasks")
			}).Order("no").First(&ws).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				// return ui.Toast("workspace-error", "warning", "", "Workspace not found!", "", nil).Render(r.Context(), w)
				return pages.NotFound().Render(r.Context(), w)
			}
			authorized := false
			userRole := ""
			for _, member := range ws.Members {
				if member.UserID == user.ID {
					authorized = true
					userRole = member.Role
					break
				}
			}
			for _, project := range ws.Projects {
				for _, member := range project.Members {
					if member.UserID == user.ID {
						authorized = true
						if userRole == "" {
							userRole = member.Role
						}
						break
					}
				}
			}
			if !authorized {
				return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
			}

			var contributors []models.User
			db.PgSql.Where("id IN (SELECT user_id FROM project_members WHERE role='CONTRIBUTOR' AND project_id IN (SELECT id FROM projects WHERE workspace_id=?))", ws.ID).Find(&contributors)

			return features.Workspace(userRole, ws, contributors).Render(r.Context(), w)
		}

		var ws []models.Workspace
		db.PgSql.Where("id in (SELECT workspace_id FROM workspace_members WHERE user_id=?)", user.ID).Preload("Members").Order("no").Find(&ws)
		return features.Workspaces(ws).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleWorkspace(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
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

		services.AddLog(user.ID, "created_workspace", "Workspace", workspace.ID, map[string]any{
			"description": "created workspace " + workspace.Name,
		})

		members := []models.WorkspaceMember{}
		member := models.WorkspaceMember{}
		no = 0
		db.PgSql.Table("workspace_members").Select("max(no)").Row().Scan(&no)
		// member.ID = utils.GenerateHash(strconv.Itoa(no + 1))
		member.WorkspaceID = workspace.ID
		member.UserID = user.ID
		member.Role = "OWNER"
		member.CreatedAt = &t
		member.CreatedBy = user.ID
		member.UpdatedAt = &t
		member.UpdatedBy = user.ID

		members = append(members, member)

		var membersPayload []string
		if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
			for _, m := range membersPayload {
				members = append(members, models.WorkspaceMember{
					// ID:          utils.GenerateHash(strconv.Itoa(no + 2 + i)),
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

		workspace.ID = r.FormValue("workspace_id")

		membersPayload := []models.WorkspaceMember{}
		if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
			for _, m := range membersPayload {
				var member = models.WorkspaceMember{}
				db.PgSql.Where("user_id=? and workspace_id=?", m.UserID, workspace.ID).First(&member)
				member.Role = m.Role
				db.PgSql.Save(&member)
			}

		}

		user, _ := auth.GetAuth(w, r)
		t := time.Now()
		if err := db.PgSql.Where("id=?", workspace.ID).First(&workspace).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "Workspace is not found!", "", nil).Render(r.Context(), w)
		}

		var member models.WorkspaceMember
		if err := db.PgSql.Where("workspace_id=? AND user_id=?", workspace.ID, user.ID).First(&member).Error; err != nil {
			println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

		updated := false
		logDetails := map[string]any{
			"description": "updated workspace",
		}

		if strings.Trim(workspace.Name, " ") != strings.Trim(r.FormValue("name"), " ") {
			logDetails["old_name"] = workspace.Name
			logDetails["new_name"] = r.FormValue("name")
			updated = true
		}
		workspace.Name = r.FormValue("name")

		if strings.Trim(workspace.Description, " ") != strings.Trim(r.FormValue("description"), " ") {
			logDetails["old_description"] = workspace.Description
			logDetails["new_description"] = r.FormValue("description")
			updated = true
		}
		workspace.Description = r.FormValue("description")

		if strings.Trim(workspace.Color, " ") != strings.Trim(r.FormValue("color"), " ") {
			logDetails["old_color"] = workspace.Color
			logDetails["new_color"] = r.FormValue("color")
			updated = true
		}
		workspace.Color = r.FormValue("color")

		if len(strings.Trim(workspace.Name, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "Name is required!", "", nil).Render(r.Context(), w)
		}

		if len(strings.Trim(workspace.Color, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "Color is required!", "", nil).Render(r.Context(), w)
		}

		if strings.ToUpper(strings.Trim(member.Role, " ")) != "ADMIN" && strings.ToUpper(strings.Trim(member.Role, " ")) != "OWNER" {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

		workspace.UpdatedAt = &t
		workspace.UpdatedBy = user.ID

		if err := db.PgSql.Save(&workspace).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("workspace-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		if updated {
			services.AddLog(user.ID, "updated_workspace", "Workspace", workspace.ID, logDetails)
		}

		// membersPayload := []models.WorkspaceMember{}
		// if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
		// 	members := ""
		// 	for i, m := range membersPayload {
		// 		if i != 0 {
		// 			members += ","
		// 		}
		// 		members += "'" + strings.Trim(m.UserID, " ") + "'"
		// 		membersPayload[i].WorkspaceID = workspace.ID
		// 		membersPayload[i].UpdatedAt = &t
		// 		membersPayload[i].UpdatedBy = user.ID
		// 	}
		// 	rawSql := fmt.Sprintf("DELETE FROM workspace_members WHERE workspace_id='%s' AND user_id NOT IN (%s) ", workspace.ID, members)
		// 	db.PgSql.Raw(rawSql)
		// 	db.PgSql.Save(&membersPayload)
		// }

		return ui.Toast("workspace-success", "success", "", "Workspace updated successfully.", "", nil).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleInviteWorkspace(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodPost:
		var workspace models.Workspace
		r.ParseForm()

		workspace.ID = r.FormValue("workspace_id")

		user, _ := auth.GetAuth(w, r)
		t := time.Now()
		if err := db.PgSql.Where("id=?", workspace.ID).First(&workspace).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "Workspace is not found!", "", nil).Render(r.Context(), w)
		}

		var member models.WorkspaceMember
		if err := db.PgSql.Where("workspace_id=? AND user_id=?", workspace.ID, user.ID).First(&member).Error; err != nil {
			println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

		updated := false
		logDetails := map[string]any{
			"description": "invited workspace",
		}

		appUrl := os.Getenv("APP_URL")
		if os.Getenv("APP_ENV") == "dev" {
			appUrl += os.Getenv("APP_PORT")
		}
		newInviteMembers := []models.InviteWorkspaceMember{}
		var inviteMembersPayload []string
		newMembers := []string{}
		if err := json.Unmarshal([]byte(r.FormValue("invite_members")), &inviteMembersPayload); err == nil {
			for _, m := range inviteMembersPayload {
				newInviteMembers = append(newInviteMembers, models.InviteWorkspaceMember{
					WorkspaceID: workspace.ID,
					Role:        "MEMBER",
					UserID:      m,
					Status:      "INVITED",
					CreatedAt:   &t,
					CreatedBy:   user.ID,
					UpdatedAt:   &t,
					UpdatedBy:   user.ID,
				})
				newMembers = append(newMembers, m)
				updated = true
			}
			logDetails["new_members"] = newMembers
		}

		if err := db.PgSql.Create(&newInviteMembers).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("workspace-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		if updated && len(newMembers) > 0 {
			services.AddLog(user.ID, "invited_workspace", "Workspace", workspace.ID, logDetails)
			for _, m := range newMembers {
				var user models.User
				db.PgSql.Where("id=?", m).First(&user)
				utils.SendInviteWorkspaceEmail(user, workspace, appUrl+"/join-workspace?id="+workspace.ID)
			}
		}

		return ui.Toast("workspace-success", "success", "", "Invite Workspace's members successfully.", "", nil).Render(r.Context(), w)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleJoinWorkspace(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetJwtClaims(w, r)
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var ws models.Workspace
			if err := db.PgSql.Where("id=?", id).
				Preload("Members", func(db *gorm.DB) *gorm.DB {
					return db.Preload("User")
				}).First(&ws).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return layouts.Layout("404 Not Found", user, pages.NotFound()).Render(r.Context(), w)
			}

			for _, member := range ws.Members {
				if member.UserID == user.ID {
					db.PgSql.Where("id=?", id).
						Preload("Members", func(db *gorm.DB) *gorm.DB {
							return db.Preload("User")
						}).Preload("Projects", func(db *gorm.DB) *gorm.DB {
						return db.Preload("Members").Preload("Tasks")
					}).Order("no").First(&ws)

					userRole := ""
					for _, member := range ws.Members {
						if member.UserID == user.ID {
							userRole = member.Role
							break
						}
					}
					for _, project := range ws.Projects {
						for _, member := range project.Members {
							if member.UserID == user.ID {
								if userRole == "" {
									userRole = member.Role
								}
								break
							}
						}
					}

					var contributors []models.User
					db.PgSql.Where("id IN (SELECT user_id FROM project_members WHERE role='CONTRIBUTOR' AND project_id IN (SELECT id FROM projects WHERE workspace_id=?))", ws.ID).Find(&contributors)

					return layouts.Layout("Workspace", user, features.Workspace(userRole, ws, contributors)).Render(r.Context(), w)
				}
			}

			var inviteMember models.InviteWorkspaceMember
			if err := db.PgSql.Where("workspace_id=? AND user_id=? AND status='INVITED'", ws.ID, user.ID).
				First(&inviteMember).Error; err == nil {
				return layouts.Layout("Workspace", user, features.JoinWorkspace(ws)).Render(r.Context(), w)
			}

			return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
		}

		return layouts.Layout("404 Not Found", user, pages.NotFound()).Render(r.Context(), w)
	case http.MethodPost:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var ws models.Workspace
			if err := db.PgSql.Where("id=?", id).
				Preload("Members", func(db *gorm.DB) *gorm.DB {
					return db.Preload("User")
				}).First(&ws).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return pages.NotFound().Render(r.Context(), w)
			}

			for _, member := range ws.Members {
				if member.UserID == user.ID {
					db.PgSql.Where("id=?", id).
						Preload("Members", func(db *gorm.DB) *gorm.DB {
							return db.Preload("User")
						}).Preload("Projects", func(db *gorm.DB) *gorm.DB {
						return db.Where("id IN (SELECT project_id FROM project_members WHERE user_id=?)", user.ID).
							Preload("Members").Preload("Tasks")
					}).Order("no").First(&ws)

					userRole := ""
					for _, member := range ws.Members {
						if member.UserID == user.ID {
							userRole = member.Role
							break
						}
					}
					for _, project := range ws.Projects {
						for _, member := range project.Members {
							if member.UserID == user.ID {
								if userRole == "" {
									userRole = member.Role
								}
								break
							}
						}
					}

					var contributors []models.User
					db.PgSql.Where("id IN (SELECT user_id FROM project_members WHERE role='CONTRIBUTOR' AND project_id IN (SELECT id FROM projects WHERE workspace_id=?))", ws.ID).Find(&contributors)

					return features.Workspace(userRole, ws, contributors).Render(r.Context(), w)
				}
			}

			var inviteMember models.InviteWorkspaceMember
			if err := db.PgSql.Where("workspace_id=? AND user_id=? AND status='INVITED'", ws.ID, user.ID).
				First(&inviteMember).Error; err == nil {
				return features.JoinWorkspace(ws).Render(r.Context(), w)
			}

			return pages.Forbidden().Render(r.Context(), w)
		}

		r.ParseForm()
		t := time.Now().Local()
		var inviteMember models.InviteWorkspaceMember
		if err := db.PgSql.Where("workspace_id=? AND user_id=? AND status='INVITED'", r.FormValue("workspace_id"), user.ID).
			First(&inviteMember).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("workspace-error", "warning", "", "You're not authorized to join this workspace!", "", nil).Render(r.Context(), w)
		}
		if r.FormValue("action") == "join" {
			inviteMember.Status = "JOINED"
			inviteMember.UpdatedAt = &t
			inviteMember.UpdatedBy = user.ID
			if err := db.PgSql.Save(&inviteMember).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return ui.Toast("workspace-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
			}
			workspaceMember := models.WorkspaceMember{
				WorkspaceID: inviteMember.WorkspaceID,
				UserID:      inviteMember.UserID,
				Role:        "MEMBER",
				CreatedAt:   &t,
				CreatedBy:   user.ID,
				UpdatedAt:   &t,
				UpdatedBy:   user.ID,
			}
			if err := db.PgSql.Create(&workspaceMember).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return ui.Toast("workspace-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
			}
			services.AddLog(user.ID, "joined_workspace", "Workspace", inviteMember.WorkspaceID, map[string]any{
				"description": "joined workspace",
			})
			return ui.Toast("workspace-success", "success", "", "You have successfully joined the workspace.", "", nil).Render(r.Context(), w)
		} else if r.FormValue("action") == "decline" {
			inviteMember.Status = "DECLINED"
			inviteMember.UpdatedAt = &t
			inviteMember.UpdatedBy = user.ID
			if err := db.PgSql.Save(&inviteMember).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return ui.Toast("workspace-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
			}
			services.AddLog(user.ID, "declined_workspace", "Workspace", inviteMember.WorkspaceID, map[string]any{
				"description": "declined workspace",
			})
			return ui.Toast("workspace-error", "error", "", "You have declined to join the workspace.", "", nil).Render(r.Context(), w)
		}

		return ui.Toast("workspace-error", "warning", "", "Invalid action!", "", nil).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
