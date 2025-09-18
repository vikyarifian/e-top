package handlers

import (
	"encoding/json"
	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/services"
	"etop/templates/components/ui"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
	"etop/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

func HandleCreateTaskFrom(w http.ResponseWriter, r *http.Request) error {
	project_id := r.URL.Query().Get("project_id")
	typ := r.URL.Query().Get("type")
	return features.CreateTaskForm(project_id, string(typ)).Render(r.Context(), w)
}

func HandleTasks(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var task models.Task
			if err := db.PgSql.Where("id=?", id).Preload("Assignees", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Preload("Watchers", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Preload("Project").First(&task).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return layouts.Layout("404 Not Found", user, pages.NotFound()).Render(r.Context(), w)
			}

			return layouts.Layout("Task", user, features.Task(task, user)).Render(r.Context(), w)
		}
		var tasks []models.Task
		db.PgSql.Where("id in (SELECT task_id FROM task_assignees WHERE user_id=?)", user.ID).
			Preload("Assignees", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Preload("Watchers", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User")
		}).Find(&tasks)
		return layouts.Layout("Tasks", user, features.Tasks(tasks)).Render(r.Context(), w)
	case http.MethodPost:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var task models.Task
			if err := db.PgSql.Where("id=?",
				id).Preload("Assignees", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Preload("Watchers", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Preload("Project").First(&task).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return pages.NotFound().Render(r.Context(), w)
			}

			return features.Task(task, user).Render(r.Context(), w)
		}

		var tasks []models.Task
		db.PgSql.Where("id in (SELECT task_id FROM task_assignees WHERE user_id=?)", user.ID).Preload("Assignees").Preload("Watchers").Order("no").Find(&tasks)
		return features.Tasks(tasks).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleTask(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodPost:
		var task models.Task
		r.ParseForm()
		t := time.Now().Local()
		user, _ := auth.GetAuth(w, r)

		task.Title = r.FormValue("title")
		if len(strings.Trim(task.Title, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Title is required!", "", nil).Render(r.Context(), w)
		}

		task.Description = r.FormValue("description")
		task.ProjectID = r.FormValue("project_id")
		task.Type = r.FormValue("type")

		if strings.Trim(task.Type, " ") == "PROJECT" {
			if len(strings.Trim(task.ProjectID, " ")) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("task-error", "warning", "", "Project is required!", "", nil).Render(r.Context(), w)
			}

			project := models.Project{}
			if err := db.PgSql.Where("id=?", task.ProjectID).First(&project).Error; err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("task-error", "warning", "", "Project is not found!", "", nil).Render(r.Context(), w)
			}

			projectStatuses := services.GetProjectStatuses()
			for _, projectStatus := range projectStatuses {
				if strings.EqualFold(strings.Trim(projectStatus.Status, " "), strings.Trim(project.Status, " ")) {
					if projectStatus.Value > 4 || projectStatus.Value == 0 {
						w.WriteHeader(http.StatusBadRequest)
						return ui.Toast("task-error", "warning", "", "Cannot add task to a completed or cancelled project!", "", nil).Render(r.Context(), w)
					}
					break
				}
			}

			var member models.ProjectMember
			if err := db.PgSql.Where("project_id=? AND user_id=?", task.ProjectID, user.ID).First(&member).Error; err != nil {
				w.WriteHeader(http.StatusForbidden)
				return ui.Toast("task-error", "warning", "", "You are not a member of this project!", "", nil).Render(r.Context(), w)
			}
			if strings.ToUpper(strings.Trim(member.Role, " ")) == "VIEWER" || strings.ToUpper(strings.Trim(member.Role, " ")) == "CONTRIBUTOR" {
				w.WriteHeader(http.StatusForbidden)
				return ui.Toast("task-error", "warning", "", "You don't have permission to add task to this project!", "", nil).Render(r.Context(), w)
			}
		} else {
			if strings.Trim(task.Type, " ") == "" {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("task-error", "warning", "", "Invalid task type!", "", nil).Render(r.Context(), w)
			}
		}

		task.Status = r.FormValue("status")
		if len(strings.Trim(task.Status, " ")) < 1 {
			task.Status = "TO_DO"
		}

		task.StartDate = &t

		taskStatus := services.GetTaskStatuses()
		validStatus := false
		statusDone := false
		for _, s := range taskStatus {
			if strings.EqualFold(strings.Trim(s.Status, " "), strings.Trim(task.Status, " ")) {
				validStatus = true
				if s.Value == 5 {
					statusDone = true
					tc := time.Now().Add(1 * time.Minute)
					task.CompletedAt = &tc
					task.ActualHours = float32(utils.TimeDiff(*task.StartDate, tc).Hours())
				}
				break
			}
		}
		if !validStatus {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Invalid status!", "", nil).Render(r.Context(), w)
		}

		dueDate, err := time.Parse("2006-01-02", r.FormValue("due_date"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Invalid due date format!", "", nil).Render(r.Context(), w)
		}
		task.DueDate = &dueDate

		task.EstimatedHours = float32(utils.TimeDiff(*task.StartDate, *task.DueDate).Hours())

		task.Priority = r.FormValue("priority")
		if len(strings.Trim(task.Priority, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Priority is required!", "", nil).Render(r.Context(), w)
		}

		taskPriorities := services.GetTaskPriorities()
		validPriority := false
		for _, p := range taskPriorities {
			if p.Priority == task.Priority {
				validPriority = true
				break
			}
		}
		if !validPriority {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Invalid priority!", "", nil).Render(r.Context(), w)
		}

		if dueDate.Before(*task.StartDate) {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Due date cannot be before today!", "", nil).Render(r.Context(), w)
		}

		no := 0
		db.PgSql.Table("tasks").Select("max(no)").Row().Scan(&no)
		task.ID = utils.GenerateHash(strconv.Itoa(no + 1))
		task.CreatedAt = &t
		task.CreatedBy = user.ID
		task.UpdatedAt = &t
		task.UpdatedBy = user.ID

		if err := db.PgSql.Create(&task).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("task-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		assignees := []models.TaskAssignee{}
		assigneesPayload := []string{}
		if err := json.Unmarshal([]byte(r.FormValue("assignees")), &assigneesPayload); err == nil {
			for _, m := range assigneesPayload {
				assignee := models.TaskAssignee{
					TaskID:    task.ID,
					UserID:    m,
					Status:    task.Status,
					CreatedAt: &t,
					CreatedBy: user.ID,
					UpdatedAt: &t,
					UpdatedBy: user.ID,
				}
				if statusDone {
					assignee.ActualHours = task.ActualHours
					assignee.CompletedAt = task.CompletedAt
				}
				assignees = append(assignees, assignee)
			}
		}

		db.PgSql.Create(&assignees)

		return ui.Toast("task-success", "success", "", "Task created successfully.", "", nil).Render(r.Context(), w)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleTaskWatcher(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		taskID := r.URL.Query().Get("task_id")
		var watchers []models.TaskWatchers
		if err := db.PgSql.Where("task_id=?", taskID).Preload("User").Find(&watchers).Error; err != nil {
			w.WriteHeader(http.StatusOK)
			return features.TaskWatchers([]models.TaskWatchers{}).Render(r.Context(), w)
		}
		println(len(watchers))
		w.WriteHeader(http.StatusOK)
		return features.TaskWatchers(watchers).Render(r.Context(), w)
	case http.MethodPost:
		taskID := r.URL.Query().Get("task_id")
		userID := r.URL.Query().Get("user_id")
		status := r.URL.Query().Get("status")
		if strings.Trim(taskID, " ") == "" || strings.Trim(userID, " ") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-watcher-error", "warning", "", "Task and User are required!", "", nil).Render(r.Context(), w)
		}
		var task models.Task
		if err := db.PgSql.Where("id=?", taskID).Preload("Watchers").First(&task).Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			return ui.Toast("task-watcher-error", "warning", "", "Task not found!", "", nil).Render(r.Context(), w)
		}
		var user models.User
		if err := db.PgSql.Where("id=?", userID).First(&user).Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			return ui.Toast("task-watcher-error", "warning", "", "User not found!", "", nil).Render(r.Context(), w)
		}
		t := time.Now().Local()
		if strings.Trim(status, " ") == "true" {
			if services.IsUserWatchingTask(task, userID) {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("task-watcher-warning", "warning", "", "You're already watching this task!", "", nil).Render(r.Context(), w)
			}
			watcher := models.TaskWatchers{
				TaskID:    task.ID,
				UserID:    user.ID,
				CreatedAt: &t,
				CreatedBy: user.ID,
				UpdatedAt: &t,
				UpdatedBy: user.ID,
			}
			if err := db.PgSql.Create(&watcher).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return ui.Toast("task-watcher-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
			}
			return ui.Toast("task-watcher-success", "success", "", "Watchers added successfully.", "", nil).Render(r.Context(), w)
		}
		if strings.Trim(status, " ") == "false" {
			if !services.IsUserWatchingTask(task, userID) {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("task-watcher-warning", "warning", "", "You're' not watching this task!", "", nil).Render(r.Context(), w)
			}
			if err := db.PgSql.Where("task_id=? AND user_id=?", task.ID, user.ID).Delete(&models.TaskWatchers{}).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return ui.Toast("task-watcher-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
			}
			return ui.Toast("task-watcher-success", "success", "", "Watchers removed successfully.", "", nil).Render(r.Context(), w)
		}
		w.WriteHeader(http.StatusBadRequest)
		return ui.Toast("task-watcher-error", "warning", "", "Invalid status value!", "", nil).Render(r.Context(), w)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
