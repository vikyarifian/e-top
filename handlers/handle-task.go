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
			if err := db.PgSql.Where("id=?", id).Preload("Assignee", func(db *gorm.DB) *gorm.DB {
				return db //.Preload("User")
			}).Preload("Watchers", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Preload("Project", func(db *gorm.DB) *gorm.DB {
				return db.Preload("Members")
			}).Preload("Status").Preload("Priority").First(&task).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return layouts.Layout("404 Not Found", user, pages.NotFound()).Render(r.Context(), w)
			}
			authorized := false
			for _, member := range task.Project.Members {
				if member.UserID == user.ID {
					authorized = true
					break
				}
			}
			if task.Assignee.ID == user.ID || task.CreatedBy == user.ID {
				authorized = true
			}
			if !authorized {
				return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
			}
			return layouts.Layout("Task", user, features.Task(task, user)).Render(r.Context(), w)
		}
		var tasks []models.Task
		db.PgSql.Where("id in (SELECT task_id FROM task_assignees WHERE user_id=?)", user.ID).
			Preload("Assignee", func(db *gorm.DB) *gorm.DB {
				return db //.Preload("User")
			}).Preload("Watchers", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User")
		}).Preload("Status").Preload("Priority").Find(&tasks)
		return layouts.Layout("Tasks", user, features.Tasks(tasks)).Render(r.Context(), w)
	case http.MethodPost:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var task models.Task
			if err := db.PgSql.Where("id=?",
				id).Preload("Assignee", func(db *gorm.DB) *gorm.DB {
				return db //.Preload("User")
			}).Preload("Watchers", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Preload("Project", func(db *gorm.DB) *gorm.DB {
				return db.Preload("Members")
			}).Preload("Status").Preload("Priority").First(&task).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return pages.NotFound().Render(r.Context(), w)
			}
			authorized := false
			for _, member := range task.Project.Members {
				if member.UserID == user.ID {
					authorized = true
					break
				}
			}
			if task.Assignee.ID == user.ID || task.CreatedBy == user.ID {
				authorized = true
			}
			if !authorized {
				return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
			}
			return features.Task(task, user).Render(r.Context(), w)
		}

		var tasks []models.Task
		db.PgSql.Where("id in (SELECT task_id FROM task_assignees WHERE user_id=?)", user.ID).Preload("Assignee", func(db *gorm.DB) *gorm.DB {
			return db //.Preload("User")
		}).Preload("Watchers").Order("no").Preload("Status").Preload("Priority").Find(&tasks)
		return features.Tasks(tasks).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleUpdateTask(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodPost:
		var task models.Task
		var taskStatus models.TaskStatus
		r.ParseForm()
		t := time.Now().Local()
		user, _ := auth.GetAuth(w, r)

		if err := db.PgSql.Where("id=?", r.FormValue("id")).First(&task).Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			return ui.Toast("task-error", "warning", "", "Task not found!", "", nil).Render(r.Context(), w)
		}

		oldStatusID := task.StatusID

		if task.CreatedBy != user.ID {
			w.WriteHeader(http.StatusForbidden)
			return ui.Toast("task-error", "warning", "", "You don't have permission to update this task!", "", nil).Render(r.Context(), w)
		}

		if err := db.PgSql.Where("no=?", r.FormValue("status_id")).First(&taskStatus).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Invalid status!", "", nil).Render(r.Context(), w)
		}

		task.StatusID = taskStatus.No
		task.StatusLabel = taskStatus.Status
		task.UpdatedAt = &t
		task.UpdatedBy = user.ID

		logDescription := "updated task status to " + taskStatus.Label
		if taskStatus.Value == 5 {
			tc := time.Now().Add(1 * time.Minute)
			task.CompletedAt = &tc
			if task.StartDate != nil {
				task.ActualHours = float32(utils.TimeDiff(*task.StartDate, tc).Hours())
			} else {
				task.ActualHours = 0
			}
			logDescription += " and marked it as completed"
		} else {
			task.CompletedAt = nil
			task.ActualHours = 0
		}

		if err := db.PgSql.Save(&task).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("task-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}
		services.AddLog(user.ID, "updated_task", "Task", task.ID, map[string]any{
			"description":   logDescription,
			"new_status_id": taskStatus.No,
			"old_status_id": oldStatusID,
		})

		db.PgSql.Where("id=?", task.ID).Preload("Assignee", func(db *gorm.DB) *gorm.DB {
			return db //.Preload("User")
		}).Preload("Watchers", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User")
		}).Preload("Project", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Members")
		}).Preload("Status").Preload("Priority").First(&task)
		return features.Task(task, user).Render(r.Context(), w)
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

		statusIDStr := r.FormValue("status_id")
		taskStatus := services.GetTaskStatuses()
		if strings.Trim(statusIDStr, " ") == "" || statusIDStr == "0" {
			for _, s := range taskStatus {
				if s.Level == 1 && s.Form == 1 {
					statusIDStr = strconv.Itoa(s.No)
				}
			}
			// w.WriteHeader(http.StatusBadRequest)
			// return ui.Toast("task-error", "warning", "", "Status is required!", "", nil).Render(r.Context(), w)
		}
		sid, err := strconv.Atoi(statusIDStr)
		if err != nil {

			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Invalid status id!", "", nil).Render(r.Context(), w)
		}
		task.StatusID = sid

		task.StartDate = &t

		validStatus := false
		statusDone := false

		for _, s := range taskStatus {
			if s.No == task.StatusID {
				validStatus = true
				task.StatusLabel = s.Status
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

		priorityValue := strings.TrimSpace(r.FormValue("priority_id"))
		if len(priorityValue) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Priority is required!", "", nil).Render(r.Context(), w)
		}

		priorityID, err := strconv.Atoi(priorityValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Invalid priority id!", "", nil).Render(r.Context(), w)
		}
		task.PriorityID = priorityID

		taskPriorities := services.GetTaskPriorities()
		validPriority := false
		for _, p := range taskPriorities {
			if p.No == task.PriorityID {
				validPriority = true
				task.PriorityLabel = p.Priority
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

		task.UserID = r.FormValue("assignee")
		if len(strings.Trim(task.UserID, " ")) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("task-error", "warning", "", "Assignee is required!", "", nil).Render(r.Context(), w)
		}

		no := 0
		db.PgSql.Table("tasks").Select("max(no)").Row().Scan(&no)
		task.ID = utils.GenerateHash(strconv.Itoa(no + 1))
		task.CreatedAt = &t
		task.CreatedBy = user.ID
		task.UpdatedAt = &t
		task.UpdatedBy = user.ID
		if statusDone {
			task.ActualHours = task.ActualHours
			task.CompletedAt = task.CompletedAt
		}

		if task.Type != "PROJECT" {
			if task.Assignee.ID != user.ID {
				task.Type = "TICKET"
			} else {
				task.Type = "DAILY"
			}
		}

		if err := db.PgSql.Create(&task).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("task-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		services.AddLog(user.ID, "created_task", "Task", task.ID, map[string]any{
			"description": "created task " + task.Title,
		})

		// assignees := []models.TaskAssignee{}
		// assigneesPayload := []string{}
		// if err := json.Unmarshal([]byte(r.FormValue("assignees")), &assigneesPayload); err == nil {
		// 	for _, m := range assigneesPayload {
		// 		assignee := models.TaskAssignee{
		// 			TaskID:    task.ID,
		// 			UserID:    m,
		// 			Status:    task.Status,
		// 			CreatedAt: &t,
		// 			CreatedBy: user.ID,
		// 			UpdatedAt: &t,
		// 			UpdatedBy: user.ID,
		// 		}
		// 		if statusDone {
		// 			assignee.ActualHours = task.ActualHours
		// 			assignee.CompletedAt = task.CompletedAt
		// 		}
		// 		assignees = append(assignees, assignee)
		// 	}
		// }

		// db.PgSql.Create(&assignees)

		tags := []models.TaskTag{}
		var tagsPayload []string
		if err := json.Unmarshal([]byte(r.FormValue("tags")), &tagsPayload); err == nil {
			for _, tag := range tagsPayload {
				tags = append(tags, models.TaskTag{
					TaskID: task.ID,
					Tag:    tag,
				})
			}
		}

		db.PgSql.Create(&tags)

		return ui.Toast("task-success", "success", "", "Task created successfully.", "", nil).Render(r.Context(), w)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleEditTask(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodPut:
		user, _ := auth.GetAuth(w, r)
		r.ParseForm()
		t := time.Now().Local()

		taskID := r.FormValue("task_id")
		var task models.Task
		if err := db.PgSql.Where("id=?", taskID).First(&task).Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			return ui.Toast("edit-task-error", "warning", "", "Task not found!", "", nil).Render(r.Context(), w)
		}

		canEdit := task.CreatedBy == user.ID || user.Level == "ADMIN"
		if !canEdit && task.ProjectID != "" {
			var member models.ProjectMember
			if err := db.PgSql.Where("project_id=? AND user_id=? AND role IN ('ADMIN','OWNER')", task.ProjectID, user.ID).First(&member).Error; err == nil {
				canEdit = true
			}
		}
		if !canEdit {
			w.WriteHeader(http.StatusForbidden)
			return ui.Toast("edit-task-error", "warning", "", "You don't have permission to edit this task!", "", nil).Render(r.Context(), w)
		}

		logDetails := map[string]any{
			"description": "updated task",
		}

		title := strings.TrimSpace(r.FormValue("title"))
		if title == "" {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("edit-task-error", "warning", "", "Title is required!", "", nil).Render(r.Context(), w)
		}
		if task.Title != title {
			logDetails["old_title"] = task.Title
			logDetails["new_title"] = title
		}
		task.Title = title

		desc := r.FormValue("description")
		if task.Description != desc {
			logDetails["old_description"] = task.Description
			logDetails["new_description"] = desc
		}
		task.Description = desc

		statusIDStr := r.FormValue("status_id")
		if statusIDStr != "" {
			sid, _ := strconv.Atoi(statusIDStr)
			var taskStatus models.TaskStatus
			if err := db.PgSql.Where("no=?", sid).First(&taskStatus).Error; err == nil {
				if task.StatusID != taskStatus.No {
					logDetails["old_status_id"] = task.StatusID
					logDetails["new_status_id"] = taskStatus.No
				}
				task.StatusID = taskStatus.No
				task.StatusLabel = taskStatus.Status
			}
		}

		priorityIDStr := r.FormValue("priority_id")
		if priorityIDStr != "" {
			pid, _ := strconv.Atoi(priorityIDStr)
			var taskPriority models.TaskPriority
			if err := db.PgSql.Where("no=?", pid).First(&taskPriority).Error; err == nil {
				if task.PriorityID != taskPriority.No {
					logDetails["old_priority_id"] = task.PriorityID
					logDetails["new_priority_id"] = taskPriority.No
				}
				task.PriorityID = taskPriority.No
				task.PriorityLabel = taskPriority.Priority
			}
		}

		dueDateStr := r.FormValue("due_date")
		if dueDateStr != "" {
			parsed, err := time.Parse("2006-01-02", dueDateStr)
			if err == nil {
				task.DueDate = &parsed
			}
		} else {
			task.DueDate = nil
		}

		assigneeID := r.FormValue("assignee")
		if assigneeID != "" {
			task.UserID = assigneeID
		}

		task.UpdatedAt = &t
		task.UpdatedBy = user.ID
		if task.Type != "PROJECT" {
			if task.Assignee.ID != user.ID {
				task.Type = "TICKET"
			} else {
				task.Type = "DAILY"
			}
		}
		if err := db.PgSql.Save(&task).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("edit-task-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		services.AddLog(user.ID, "updated_task", "Task", task.ID, logDetails)

		return ui.Toast("edit-task-success", "success", "", "Task updated successfully.", "", nil).Render(r.Context(), w)

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

			services.AddLog(user.ID, "watched_task", "Task", task.ID, map[string]any{
				"description": "started watching task",
			})

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

			services.AddLog(user.ID, "unwatched_task", "Task", task.ID, map[string]any{
				"description": "stopped watching task",
			})

			return ui.Toast("task-watcher-success", "success", "", "Watchers removed successfully.", "", nil).Render(r.Context(), w)
		}
		w.WriteHeader(http.StatusBadRequest)
		return ui.Toast("task-watcher-error", "warning", "", "Invalid status value!", "", nil).Render(r.Context(), w)

	case http.MethodPut:
		return nil
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
