package services

import (
	"time"

	"etop/db"
	"etop/models"
)

type DashboardData struct {
	WorkspaceCount int64
	ProjectCount   int64
	TaskCount      int64
	DoneCount      int64
	MemberCount    int64

	TCR float64
	OTR float64
	TPS float64
	WER float64

	StatusDistribution []StatusCount
	TypeDistribution   []TypeCount
	MonthlyCompletion  []MonthlyCount
}

type StatusCount struct {
	Label string
	Color string
	Count int64
}

type TypeCount struct {
	Type  string
	Count int64
}

type MonthlyCount struct {
	Year  int
	Month int
	Count int64
}

func GetDashboardData(userID string) DashboardData {
	var d DashboardData

	db.PgSql.Model(&models.Workspace{}).
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspace_members.user_id = ?", userID).
		Count(&d.WorkspaceCount)

	db.PgSql.Model(&models.Project{}).
		Joins("JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userID).
		Count(&d.ProjectCount)

	db.PgSql.Model(&models.Task{}).
		Where("(user_id = ? OR created_by = ?) AND completed_at IS NOT NULL", userID, userID).
		Count(&d.DoneCount)

	db.PgSql.Model(&models.Task{}).
		Where("user_id = ? OR created_by = ?", userID, userID).
		Count(&d.TaskCount)

	db.PgSql.Raw(`
		SELECT COUNT(DISTINCT wm.user_id)
		FROM workspace_members wm
		JOIN workspace_members wm2 ON wm2.workspace_id = wm.workspace_id
		WHERE wm2.user_id = ?
	`, userID).Scan(&d.MemberCount)

	if d.TaskCount > 0 {
		d.TCR = float64(d.DoneCount) / float64(d.TaskCount) * 100
	}

	var onTimeCount int64
	db.PgSql.Model(&models.Task{}).
		Where("(user_id = ? OR created_by = ?) AND completed_at IS NOT NULL AND completed_at <= due_date", userID, userID).
		Count(&onTimeCount)
	if d.DoneCount > 0 {
		d.OTR = float64(onTimeCount) / float64(d.DoneCount) * 100
	}

	type PriorityWeight struct {
		TotalWeight float64
	}
	var allWeight PriorityWeight
	db.PgSql.Raw(`
		SELECT COALESCE(SUM(tp.level), 0) as total_weight
		FROM tasks t
		JOIN task_priorities tp ON tp.no = t.priority_id
		WHERE t.user_id = ? OR t.created_by = ?
	`, userID, userID).Scan(&allWeight)

	var doneWeight PriorityWeight
	db.PgSql.Raw(`
		SELECT COALESCE(SUM(tp.level), 0) as total_weight
		FROM tasks t
		JOIN task_priorities tp ON tp.no = t.priority_id
		WHERE (t.user_id = ? OR t.created_by = ?) AND t.completed_at IS NOT NULL
	`, userID, userID).Scan(&doneWeight)

	if allWeight.TotalWeight > 0 {
		d.TPS = doneWeight.TotalWeight / allWeight.TotalWeight * 100
	}

	type Efficiency struct {
		Value float64
	}
	var werResult Efficiency
	db.PgSql.Raw(`
		SELECT COALESCE(AVG((t.estimated_hours / NULLIF(t.actual_hours, 0)) * 100), 0) as value
		FROM tasks t
		WHERE (t.user_id = ? OR t.created_by = ?)
			AND t.completed_at IS NOT NULL
			AND t.estimated_hours > 0
			AND t.actual_hours > 0
	`, userID, userID).Scan(&werResult)
	d.WER = werResult.Value

	db.PgSql.Raw(`
		SELECT ts.label, ts.color, COUNT(t.no) as count
		FROM tasks t
		JOIN task_statuses ts ON ts.no = t.status_id
		WHERE t.user_id = ? OR t.created_by = ?
		GROUP BY ts.label, ts.color
		ORDER BY count DESC
	`, userID, userID).Scan(&d.StatusDistribution)

	db.PgSql.Raw(`
		SELECT t.type, COUNT(t.no) as count
		FROM tasks t
		WHERE t.user_id = ? OR t.created_by = ?
		GROUP BY t.type
		ORDER BY count DESC
	`, userID, userID).Scan(&d.TypeDistribution)

	type RawMonthly struct {
		Year  int
		Month int
		Count int64
	}
	db.PgSql.Raw(`
		SELECT EXTRACT(YEAR FROM completed_at)::int as year,
			   EXTRACT(MONTH FROM completed_at)::int as month,
			   COUNT(no) as count
		FROM tasks
		WHERE (user_id = ? OR created_by = ?)
			AND completed_at IS NOT NULL
			AND completed_at >= ?
		GROUP BY year, month
		ORDER BY year, month
	`, userID, userID, time.Now().AddDate(0, -6, 0)).Scan(&d.MonthlyCompletion)

	return d
}
