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
	"etop/utils"
)

// func HandleProjects(w http.ResponseWriter, r *http.Request) error {
// 	user, _ := auth.GetAuth(w, r)
// 	switch r.Method {
// 	case http.MethodGet:

// 		return nil
// 	default:
// 		return nil
// 	}
// }

func HandleProject(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return nil
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

		project.Status = r.FormValue("status")
		if len(strings.Trim(project.Status, " ")) < 1 {
			project.Status = "PLANNING"
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

		members := []models.ProjectMember{}
		member := models.ProjectMember{}
		no = 0
		db.PgSql.Table("project_members").Select("max(no)").Row().Scan(&no)
		member.ID = utils.GenerateHash(strconv.Itoa(no + 1))
		member.ProjectID = project.ID
		member.UserID = user.ID
		member.Role = "ADMIN"
		member.CreatedAt = &t
		member.CreatedBy = user.ID
		member.UpdatedAt = &t
		member.UpdatedBy = user.ID

		members = append(members, member)

		membersPayload := []models.ProjectMember{}
		if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
			for i, m := range membersPayload {
				members = append(members, models.ProjectMember{
					ID:        utils.GenerateHash(strconv.Itoa(no + 2 + i)),
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

		project.ID = r.FormValue("id")
		if err := db.PgSql.Where("id=?", project.ID).First(&project).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "Project is not found!", "", nil).Render(r.Context(), w)
		}

		var member models.WorkspaceMember
		if err := db.PgSql.Where("id=? AND user_id=?", project.ID, user.ID).First(&member).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("project-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

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

		if strings.Trim(project.Status, " ") != "" {
			project.Status = r.FormValue("status")
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
		project.StartDate = &dueDate

		progress, err := strconv.Atoi(r.FormValue("progress"))
		if err == nil {
			project.Progress = progress
		}

		isArchivedStr := r.FormValue("is_archived")
		project.IsArchived = isArchivedStr == "true" || false
		project.UpdatedAt = &t
		project.UpdatedBy = user.ID

		if err := db.PgSql.Save(&project).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("project-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}
		return ui.Toast("project-success", "success", "", "Project updated successfully.", "", nil).Render(r.Context(), w)
	default:
		return nil
	}
}
