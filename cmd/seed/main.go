package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"etop/models"
)

var db *gorm.DB

func main() {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load(".env.local"); err != nil {
			log.Fatal(err)
		}
	}

	dsn := os.Getenv("POSTGRES_URL")
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	var user models.User
	if err := db.Where("email = ?", "vikyarifi@gmail.com").First(&user).Error; err != nil {
		log.Fatal("User vikyarifi@gmail.com not found. Register first.")
	}
	fmt.Println("User found:", user.ID, user.FullName)

	workspace := findOrCreateWorkspace(user)
	fmt.Println("Workspace:", workspace.ID)

	project := findOrCreateProject(user, workspace)
	fmt.Println("Project:", project.ID)

	force := len(os.Args) > 1 && os.Args[1] == "-force"

	seedTasks(user, project, force)
	seedComments(user, project, force)

	fmt.Println("Seed completed!")
}

func findOrCreateWorkspace(user models.User) models.Workspace {
	var ws models.Workspace
	err := db.Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspace_members.user_id = ?", user.ID).
		First(&ws).Error
	if err == nil {
		return ws
	}

	ws = models.Workspace{
		ID:          uuid.New().String(),
		Name:        "Personal Workspace",
		Description: "Workspace vikyarifi",
		Color:       "#3b82f6",
		CreatedAt:   tp(time.Now()),
		CreatedBy:   user.ID,
		UpdatedAt:   tp(time.Now()),
		UpdatedBy:   user.ID,
	}
	db.Create(&ws)

	member := models.WorkspaceMember{
		WorkspaceID: ws.ID,
		UserID:      user.ID,
		Role:        "OWNER",
		CreatedAt:   tp(time.Now()),
		CreatedBy:   user.ID,
		UpdatedAt:   tp(time.Now()),
		UpdatedBy:   user.ID,
	}
	db.Create(&member)
	return ws
}

func findOrCreateProject(user models.User, ws models.Workspace) models.Project {
	var p models.Project
	err := db.Where("title = ? AND workspace_id = ?", "General", ws.ID).First(&p).Error
	if err == nil {
		return p
	}

	p = models.Project{
		ID:          uuid.New().String(),
		Title:       "General",
		Description: "Proyek default untuk tugas umum",
		WorkspaceID: ws.ID,
		Status:      "IN_PROGRESS",
		CreatedAt:   tp(time.Now()),
		CreatedBy:   user.ID,
		UpdatedAt:   tp(time.Now()),
		UpdatedBy:   user.ID,
	}
	db.Create(&p)

	member := models.ProjectMember{
		ProjectID: p.ID,
		UserID:    user.ID,
		Role:      "OWNER",
		CreatedAt: tp(time.Now()),
		CreatedBy: user.ID,
		UpdatedAt: tp(time.Now()),
		UpdatedBy: user.ID,
	}
	db.Create(&member)
	return p
}

var seededTaskIDs []string

func seedTasks(user models.User, project models.Project, force bool) {
	var existingCount int64
	db.Model(&models.Task{}).Where("user_id = ?", user.ID).Count(&existingCount)
	if existingCount > 0 {
		if force {
			fmt.Printf("Force: deleting all %d tasks & related data...\n", existingCount)
			var taskIDs []string
			db.Model(&models.Task{}).Where("user_id = ?", user.ID).Pluck("id", &taskIDs)
			if len(taskIDs) > 0 {
				db.Where("task_id IN ?", taskIDs).Delete(&models.TaskTag{})
				db.Where("task_id IN ?", taskIDs).Delete(&models.Subtask{})
				db.Where("task_id IN ?", taskIDs).Delete(&models.Comment{})
				db.Where("resource_type = ? AND resource_id IN ?", "Task", taskIDs).Delete(&models.Log{})
				db.Where("id IN ?", taskIDs).Delete(&models.Task{})
			}
		} else {
			fmt.Printf("Tasks exist (%d). Use -force to re-seed.\n", existingCount)
			return
		}
	}

	type Status struct {
		No     int
		Status string
	}
	var statuses []Status
	db.Table("task_statuses").Order("no ASC").Find(&statuses)
	if len(statuses) < 5 {
		log.Fatal("Need at least 5 task statuses")
	}
	var priorityList []models.TaskPriority
	db.Order("no ASC").Find(&priorityList)
	if len(priorityList) == 0 {
		log.Fatal("No priorities found")
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	rng := rand.New(rand.NewSource(99))

	taskTitles := map[string][]string{
		"DAILY":   {"Morning standup & daily planning", "Code review pull requests", "Update documentation", "Fix reported bugs", "Write unit tests", "Refactor legacy module", "Database optimization", "API endpoint integration", "Team sync meeting", "Performance monitoring", "Deploy staging environment", "Backup production database"},
		"PROJECT": {"Design system architecture", "Implement authentication flow", "Build dashboard UI", "Setup CI/CD pipeline", "Create API documentation", "Develop notification service", "Database migration script", "User acceptance testing", "Load testing & benchmarking", "Security audit review", "Sprint retrospective report", "Client demo preparation"},
		"TICKET":  {"[BUG] Login page crash on mobile", "[FEATURE] Export to PDF", "[BUG] Incorrect date format", "[FEATURE] Dark mode toggle", "[BUG] Memory leak on dashboard", "[FEATURE] Batch delete tasks", "[BUG] Email not sending", "[FEATURE] Search autocomplete", "[BUG] Slow query on reports", "[FEATURE] Webhook integration", "[BUG] UI broken on Safari", "[FEATURE] Real-time notifications"},
	}

	types := []string{"DAILY", "PROJECT", "TICKET"}
	descriptions := []string{
		"Task ini perlu diselesaikan sesuai dengan prioritas yang ditentukan.",
		"Pengerjaan task ini membutuhkan koordinasi dengan tim terkait.",
		"Pastikan semua requirements terpenuhi sebelum marking sebagai done.",
		"Task ini merupakan bagian dari sprint current iteration.",
		"Perhatikan deadline yang sudah ditentukan.",
	}

	for year := 2020; year <= 2026; year++ {
		totalYearTasks := 0
		for month := 1; month <= 12; month++ {
			tasksThisMonth := 3 + rng.Intn(6)
			totalYearTasks += tasksThisMonth

			for i := 0; i < tasksThisMonth; i++ {
				taskType := types[rng.Intn(len(types))]
				titleOptions := taskTitles[taskType]
				title := titleOptions[rng.Intn(len(titleOptions))]

				day := 1 + rng.Intn(28)
				createdAt := time.Date(year, time.Month(month), day, 8+rng.Intn(8), rng.Intn(60), 0, 0, loc)

				isCompleted := rng.Float32() < 0.7

				dueDays := 3 + rng.Intn(14)
				dueDate := createdAt.AddDate(0, 0, dueDays)

				var completedAt *time.Time
				if isCompleted {
					compDays := dueDays - 3 + rng.Intn(8)
					if compDays < 1 {
						compDays = 1
					}
					compTime := createdAt.AddDate(0, 0, compDays)
					if rng.Float32() < 0.25 {
						compTime = dueDate.AddDate(0, 0, 1+rng.Intn(5))
					}
					completedAt = tp(compTime)
				}

				statusIdx := 0
				statusLabel := ""
				if isCompleted {
					statusIdx = 3
					statusLabel = statuses[3].Status
				} else {
					statusIdx = rng.Intn(3)
					statusLabel = statuses[statusIdx].Status
				}

				priorityIdx := rng.Intn(len(priorityList))
				selPri := priorityList[priorityIdx]

				estHours := 2 + rng.Intn(24)
				actHours := int(float64(estHours) * (0.6 + rng.Float64()*0.8))

				taskID := uuid.New().String()
				seededTaskIDs = append(seededTaskIDs, taskID)

				task := models.Task{
					ID:              taskID,
					Type:            taskType,
					Title:           title,
					Description:     descriptions[rng.Intn(len(descriptions))],
					ProjectID:       project.ID,
					StatusID:        statuses[statusIdx].No,
					StatusLabel:     statusLabel,
					PriorityID:      selPri.No,
					PriorityLabel:   selPri.Priority,
					UserID:          user.ID,
					DueDate:         tp(dueDate),
					CompletedAt:     completedAt,
					EstimatedHours:  float32(estHours),
					ActualHours:     float32(actHours),
					IsArchived:      false,
					CreatedAt:       tp(createdAt),
					CreatedBy:       user.ID,
					UpdatedAt:       tp(createdAt),
					UpdatedBy:       user.ID,
				}

				if err := db.Create(&task).Error; err != nil {
					fmt.Printf("Error creating task: %v\n", err)
				}
			}
		}
		fmt.Printf("  %s: %d tasks\n", fmt.Sprint(year), totalYearTasks)
	}
}

func seedComments(user models.User, project models.Project, force bool) {
	var completedTasks []models.Task
	db.Where("user_id = ? AND completed_at IS NOT NULL", user.ID).Find(&completedTasks)
	if len(completedTasks) == 0 {
		fmt.Println("  No completed tasks to comment on")
		return
	}

	var existingCount int64
	db.Model(&models.Comment{}).
		Joins("JOIN tasks ON tasks.id = comments.task_id").
		Where("tasks.user_id = ?", user.ID).
		Count(&existingCount)
	if existingCount > 0 && !force {
		fmt.Printf("  Comments already exist (%d), skipping.\n", existingCount)
		return
	}
	if force && existingCount > 0 {
		var taskIDs []string
		db.Model(&models.Task{}).Where("user_id = ?", user.ID).Pluck("id", &taskIDs)
		if len(taskIDs) > 0 {
			db.Where("task_id IN ?", taskIDs).Delete(&models.Comment{})
		}
	}

	rng := rand.New(rand.NewSource(777))

	commentTemplates := []string{
		"Task ini sudah selesai dikerjakan. Semua requirement terpenuhi.",
		"Selesai! Sudah di-review dan disetujui oleh tim.",
		"Done. Hasil testing menunjukkan semua berfungsi dengan baik.",
		"Penyelesaian task ini membutuhkan effort ekstra, tapi akhirnya selesai.",
		"Task completed. Dokumentasi sudah diupdate.",
		"Selesai sesuai deadline. Tidak ada issue yang ditemukan.",
		"Sudah di-deploy ke production. Monitoring berjalan normal.",
		"Done dengan beberapa catatan tambahan di dokumentasi.",
	}

	logActions := []string{
		"Task berhasil diselesaikan dengan hasil yang memuaskan.",
		"Pengerjaan task selesai dan sudah melalui proses review.",
		"Task telah ditandai sebagai selesai oleh assignee.",
		"Penyelesaian task dicatat dalam sistem.",
	}

	limit := len(completedTasks)
	if limit > 200 {
		limit = 200
	}

	for i := 0; i < limit; i++ {
		task := completedTasks[i]

		comment := models.Comment{
			TaskID:    task.ID,
			Text:      commentTemplates[rng.Intn(len(commentTemplates))],
			AuthorID:  user.ID,
			CreatedAt: task.CompletedAt,
			CreatedBy: user.ID,
			UpdatedAt: task.CompletedAt,
			UpdatedBy: user.ID,
		}
		if err := db.Omit("id").Create(&comment).Error; err != nil {
			fmt.Printf("  Error creating comment: %v\n", err)
		}

		if rng.Float32() < 0.5 {
			log := models.Log{
				UserID:       user.ID,
				Action:       "updated_task",
				ResourceType: "Task",
				ResourceID:   task.ID,
				Details:      models.JSONB{"message": logActions[rng.Intn(len(logActions))]},
				CreatedAt:    task.CompletedAt,
				CreatedBy:    user.ID,
				UpdatedAt:    task.CompletedAt,
				UpdatedBy:    user.ID,
			}
			if err := db.Omit("id").Create(&log).Error; err != nil {
				fmt.Printf("  Error creating log: %v\n", err)
			}
		}
	}
	fmt.Printf("  %d comments & logs created\n", limit)
}

func tp(t time.Time) *time.Time {
	return &t
}
