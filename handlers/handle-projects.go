package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/services"
	"etop/templates/components/ui"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
	"etop/utils"
)

func HandleCreateProjectForm(w http.ResponseWriter, r *http.Request) error {
	workspace_id := r.URL.Query().Get("workspace_id")
	return features.CreateProjectForm(workspace_id).Render(r.Context(), w)
}

func HandleEditProjectForm(w http.ResponseWriter, r *http.Request) error {
	project_id := r.URL.Query().Get("project_id")
	project := models.Project{}
	if err := db.PgSql.Where("id=?", project_id).Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User")
	}).Preload("Tags").First(&project).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		return pages.NotFound().Render(r.Context(), w)
	}
	return features.EditProjectForm(project).Render(r.Context(), w)
}

func HandleProjects(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetJwtClaims(w, r)
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var project models.Project
			if err := db.PgSql.Where("id=?", id).
				Preload("Members", func(db *gorm.DB) *gorm.DB {
					return db.Preload("User")
				}).Preload("Workspace").Order("no").First(&project).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return layouts.Layout("404 Not Found", user, pages.NotFound()).Render(r.Context(), w)
			}
			authorized := false
			userRole := ""
			for _, member := range project.Members {
				if member.UserID == user.ID {
					authorized = true
					userRole = member.Role
					break
				}
			}
			if !authorized {
				return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
			}
			page, perPage, sortBy, sortDir := parsePageParams(r)
			var total int64
			db.PgSql.Model(&models.Task{}).Where("project_id = ?", project.ID).Count(&total)
			var tasks []models.Task
			db.PgSql.Where("project_id = ?", project.ID).
				Preload("Assignee", func(db *gorm.DB) *gorm.DB { return db }).
				Preload("Status").Preload("Priority").
				Order(sortBy + " " + sortDir).
				Limit(perPage).Offset((page - 1) * perPage).
				Find(&tasks)
			pageInfo := models.PageInfo{
				Page:       page,
				PerPage:    perPage,
				Total:      total,
				TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
				SortBy:     sortBy,
				SortDir:    sortDir,
			}
			return layouts.Layout("Project", user, features.Project(userRole, project, tasks, pageInfo)).Render(r.Context(), w)
		}

		var projects []models.Project
		db.PgSql.Where("id in (SELECT project_id FROM project_members WHERE user_id=?)", user.ID).
			Preload("Members", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Order("no").Find(&projects)
		return layouts.Layout("Projects", user, features.Projects(projects)).Render(r.Context(), w)
	case http.MethodPost:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var project models.Project
			if err := db.PgSql.Where("id=?", id).
				Preload("Members", func(db *gorm.DB) *gorm.DB {
					return db.Preload("User")
				}).Preload("Workspace").Order("no").First(&project).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return pages.NotFound().Render(r.Context(), w)
			}
			authorized := false
			userRole := ""
			for _, member := range project.Members {
				if member.UserID == user.ID {
					authorized = true
					userRole = member.Role
					break
				}
			}
			if !authorized {
				return pages.Forbidden().Render(r.Context(), w)
			}
			page, perPage, sortBy, sortDir := parsePageParams(r)
			var total int64
			db.PgSql.Model(&models.Task{}).Where("project_id = ?", project.ID).Count(&total)
			var tasks []models.Task
			db.PgSql.Where("project_id = ?", project.ID).
				Preload("Assignee", func(db *gorm.DB) *gorm.DB { return db }).
				Preload("Status").Preload("Priority").
				Order(sortBy + " " + sortDir).
				Limit(perPage).Offset((page - 1) * perPage).
				Find(&tasks)
			pageInfo := models.PageInfo{
				Page:       page,
				PerPage:    perPage,
				Total:      total,
				TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
				SortBy:     sortBy,
				SortDir:    sortDir,
			}
			return features.Project(userRole, project, tasks, pageInfo).Render(r.Context(), w)
		}

		var projects []models.Project
		db.PgSql.Where("id in (SELECT project_id FROM project_members WHERE user_id=?)", user.ID).Preload("Members").Order("no").Find(&projects)
		return features.Projects(projects).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleProject(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodPost:
		var project models.Project
		r.ParseForm()
		t := time.Now()
		user, _ := auth.GetAuth(w, r)

		project.Title = r.FormValue("title")
		if len(strings.Trim(project.Title, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Title is required!", "", nil).Render(r.Context(), w)
		}

		project.Description = r.FormValue("description")
		project.WorkspaceID = r.FormValue("workspace_id")
		if len(strings.Trim(project.WorkspaceID, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Workspace is required!", "", nil).Render(r.Context(), w)
		}

		workspace := models.Workspace{}
		if err := db.PgSql.Where("id=?", project.WorkspaceID).First(&workspace).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Workspace is not found!", "", nil).Render(r.Context(), w)
		}

		var wsMember models.WorkspaceMember
		if err := db.PgSql.Where("workspace_id=? AND user_id=?", project.WorkspaceID, user.ID).First(&wsMember).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

		project.Status = r.FormValue("status")
		if len(strings.Trim(project.Status, " ")) < 1 {
			project.Status = "PLANNING"
		}

		projectStatuses := services.GetProjectStatuses()
		validStatus := false
		for _, s := range projectStatuses {
			if s.Status == project.Status {
				validStatus = true
				break
			}
		}
		if !validStatus {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Invalid project status!", "", nil).Render(r.Context(), w)
		}

		startDate, err := time.Parse("2006-01-02", r.FormValue("start_date"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Invalid start date format!", "", nil).Render(r.Context(), w)
		}
		project.StartDate = &startDate

		dueDate, err := time.Parse("2006-01-02", r.FormValue("due_date"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Invalid due date format!", "", nil).Render(r.Context(), w)
		}
		project.DueDate = &dueDate

		if dueDate.Before(startDate) {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Due date cannot be before start date!", "", nil).Render(r.Context(), w)
		}

		progress, err := strconv.Atoi(r.FormValue("progress"))
		if err != nil {
			project.Progress = 0
		} else {
			project.Progress = progress
		}

		no := 0
		db.PgSql.Table("projects").Select("max(no)").Row().Scan(&no)
		project.ID = utils.GenerateHash(strconv.Itoa(no + 1))
		project.IsArchived = false
		project.CreatedAt = &t
		project.CreatedBy = user.ID
		project.UpdatedAt = &t
		project.UpdatedBy = user.ID

		if err := db.PgSql.Create(&project).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("project-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		services.AddLog(user.ID, "created_project", "Project", project.ID, map[string]any{
			"description": "created project " + project.Title,
		})

		member := models.ProjectMember{}
		members := []models.ProjectMember{}

		no = 0
		db.PgSql.Table("project_members").Select("max(no)").Row().Scan(&no)
		// member.ID = utils.GenerateHash(strconv.Itoa(no + 1))
		member.ProjectID = project.ID
		member.UserID = user.ID
		member.Role = "OWNER"
		member.CreatedAt = &t
		member.CreatedBy = user.ID
		member.UpdatedAt = &t
		member.UpdatedBy = user.ID

		members = append(members, member)

		membersPayload := []models.ProjectMember{}
		if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
			for _, m := range membersPayload {
				members = append(members, models.ProjectMember{
					// ID:        utils.GenerateHash(strconv.Itoa(no + 2 + i)),
					ProjectID: project.ID,
					Role:      m.Role,
					UserID:    m.UserID,
					CreatedAt: &t,
					CreatedBy: user.ID,
					UpdatedAt: &t,
					UpdatedBy: user.ID,
				})
			}
		}

		db.PgSql.Create(&members)

		tags := []models.ProjectTag{}
		var tagsPayload []string
		if err := json.Unmarshal([]byte(r.FormValue("tags")), &tagsPayload); err == nil {
			for _, tag := range tagsPayload {
				tags = append(tags, models.ProjectTag{
					ProjectID: project.ID,
					Tag:       tag,
				})
			}
		}

		db.PgSql.Create(&tags)

		return ui.Toast("project-success", "success", "", "Project created successfully.", "", nil).Render(r.Context(), w)

	case http.MethodPut:
		var project models.Project
		r.ParseForm()
		t := time.Now()
		user, _ := auth.GetAuth(w, r)

		logDetails := map[string]any{
			"description": "updated project",
		}

		project.ID = r.FormValue("project_id")
		if err := db.PgSql.Where("id=? AND workspace_id=?", project.ID, r.FormValue("workspace_id")).First(&project).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Project is not found!", "", nil).Render(r.Context(), w)
		}

		var member models.ProjectMember
		if err := db.PgSql.Where("project_id=? AND user_id=?", project.WorkspaceID, user.ID).First(&member).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

		if member.Role != "OWNER" && member.Role != "ADMIN" {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

		if strings.Trim(project.Title, " ") != strings.Trim(r.FormValue("title"), " ") {
			logDetails["old_title"] = project.Title
			logDetails["new_title"] = r.FormValue("title")
		}
		project.Title = r.FormValue("title")
		if len(strings.Trim(project.Title, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Title is required!", "", nil).Render(r.Context(), w)
		}

		if strings.Trim(project.Description, " ") != strings.Trim(r.FormValue("description"), " ") {
			logDetails["old_description"] = project.Description
			logDetails["new_description"] = r.FormValue("description")
		}
		project.Description = r.FormValue("description")
		if strings.Trim(project.WorkspaceID, " ") != strings.Trim(r.FormValue("workspace_id"), " ") {
			logDetails["old_workspace_id"] = project.WorkspaceID
			logDetails["new_workspace_id"] = r.FormValue("workspace_id")
		}
		project.WorkspaceID = r.FormValue("workspace_id")
		if len(strings.Trim(project.WorkspaceID, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Workspace is required!", "", nil).Render(r.Context(), w)
		}

		workspace := models.Workspace{}
		if err := db.PgSql.Where("id=?", project.WorkspaceID).First(&workspace).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Workspace is not found!", "", nil).Render(r.Context(), w)
		}

		if strings.Trim(project.Status, " ") != strings.Trim(r.FormValue("status"), " ") {
			logDetails["old_status"] = project.Status
			logDetails["new_status"] = r.FormValue("status")
		}
		if strings.Trim(project.Status, " ") != "" {
			project.Status = r.FormValue("status")
		}

		projectStatuses := services.GetProjectStatuses()
		validStatus := false
		for _, s := range projectStatuses {
			if s.Status == project.Status {
				validStatus = true
				break
			}
		}
		if !validStatus {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Invalid project status!", "", nil).Render(r.Context(), w)
		}

		startDate, err := time.Parse("2006-01-02", r.FormValue("start_date"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Invalid start date format!", "", nil).Render(r.Context(), w)
		}
		if project.StartDate != &startDate {
			logDetails["old_start_date"] = project.StartDate
			logDetails["new_start_date"] = &startDate
		}
		project.StartDate = &startDate

		dueDate, err := time.Parse("2006-01-02", r.FormValue("due_date"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Invalid due date format!", "", nil).Render(r.Context(), w)
		}
		if project.DueDate != &dueDate {
			logDetails["old_due_date"] = project.DueDate
			logDetails["new_due_date"] = &dueDate
		}
		project.DueDate = &dueDate

		progress, err := strconv.Atoi(r.FormValue("progress"))
		if err == nil {
			if project.Progress != progress {
				logDetails["old_progress"] = project.Progress
				logDetails["new_progress"] = progress
			}
			project.Progress = progress
		}

		isArchivedStr := r.FormValue("is_archived")
		if (!project.IsArchived && isArchivedStr == "true") || (project.IsArchived && isArchivedStr != "true") {
			logDetails["old_is_archived"] = project.Progress
			logDetails["new_is_archived"] = isArchivedStr == "true" || false
		}
		project.IsArchived = isArchivedStr == "true" || false

		project.UpdatedAt = &t
		project.UpdatedBy = user.ID

		if err := db.PgSql.Save(&project).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("project-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		services.AddLog(user.ID, "updated_project", "Project", project.ID, logDetails)

		membersPayload := []models.ProjectMember{}
		if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
			members := ""
			for i, m := range membersPayload {
				if i != 0 {
					members += ","
				}
				members += "'" + strings.Trim(m.UserID, " ") + "'"
				mp := models.ProjectMember{}
				if err := db.PgSql.Where("project_id=? AND user_id=?", project.ID, m.UserID).First(&mp).Error; err == nil {
					mp.Role = m.Role
					mp.UpdatedAt = &t
					mp.UpdatedBy = user.ID
					db.PgSql.Save(&mp)
				} else {
					mp.ProjectID = project.ID
					mp.UserID = m.UserID
					mp.Role = m.Role
					mp.CreatedAt = &t
					mp.CreatedBy = user.ID
					mp.UpdatedAt = &t
					mp.UpdatedBy = user.ID
					db.PgSql.Create(&mp)
				}
			}

			rawSql := fmt.Sprintf("DELETE FROM project_members WHERE project_id='%s' AND user_id NOT IN (%s) AND role != 'OWNER'", project.ID, members)
			if err := db.PgSql.Exec(rawSql).Error; err != nil {
				fmt.Printf("Error executing delete project_members: %v\n", err)
				if err != nil {
					// Proper error logging
					fmt.Printf("Error deleting project members: %v\n", err)
				}
			}
		}

		tags := ""
		var tagsPayload []string
		if err := json.Unmarshal([]byte(r.FormValue("tags")), &tagsPayload); err == nil {
			for _, tag := range tagsPayload {
				pt := models.ProjectTag{}
				if err := db.PgSql.Where("project_id=? AND tag=?", project.ID, tag).First(&pt).Error; err != nil {
					pt.ProjectID = project.ID
					pt.Tag = tag
					db.PgSql.Create(&pt)
				}
				tags += "'" + strings.Trim(tag, " ") + "',"
			}
			rawSql := fmt.Sprintf("DELETE FROM project_tags WHERE project_id='%s' AND tag NOT IN (%s)", project.ID, strings.TrimRight(tags, ","))
			if err := db.PgSql.Exec(rawSql).Error; err != nil {
				fmt.Printf("Error executing delete project_tags: %v\n", err)
				if err != nil {
					fmt.Printf("Error deleting project tags: %v\n", err)
				}
			}
		}

		return ui.Toast("project-success", "success", "", "Project updated successfully.", "", nil).Render(r.Context(), w)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
