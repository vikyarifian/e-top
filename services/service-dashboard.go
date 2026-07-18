package services

import (
	"strconv"
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

type AchievedEvaluation struct {
	TCR        float64
	OTR        float64
	TPS        float64
	WER        float64
	FinalScore float64
	Category   string

	TaskCount    int64
	DoneCount    int64
	OnTimeCount  int64
	ProjectCount int64

	StatusDistribution []StatusCount
	TypeDistribution   []TypeCount
	MonthlyCompletion  []MonthlyCount
}

func GetAchievedEvaluation(userID string, year string) AchievedEvaluation {
	var e AchievedEvaluation

	doneQuery := db.PgSql.Model(&models.Task{}).
		Where("user_id = ? AND completed_at IS NOT NULL", userID)
	if year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			doneQuery = doneQuery.Where("EXTRACT(YEAR FROM completed_at) = ?", y)
		}
	}
	doneQuery.Count(&e.DoneCount)

	allQuery := db.PgSql.Model(&models.Task{}).
		Where("user_id = ?", userID)
	if year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			allQuery = allQuery.Where("EXTRACT(YEAR FROM created_at) = ?", y)
		}
	}
	allQuery.Count(&e.TaskCount)

	onTimeQuery := db.PgSql.Model(&models.Task{}).
		Where("user_id = ? AND completed_at IS NOT NULL AND completed_at <= due_date", userID)
	if year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			onTimeQuery = onTimeQuery.Where("EXTRACT(YEAR FROM completed_at) = ?", y)
		}
	}
	onTimeQuery.Count(&e.OnTimeCount)

	if e.TaskCount > 0 {
		e.TCR = float64(e.DoneCount) / float64(e.TaskCount) * 100
	}
	if e.DoneCount > 0 {
		e.OTR = float64(e.OnTimeCount) / float64(e.DoneCount) * 100
	}

	projectQuery := db.PgSql.Model(&models.Task{}).
		Where("user_id = ? AND project_id <> ''", userID)
	if year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			projectQuery = projectQuery.Where("EXTRACT(YEAR FROM created_at) = ?", y)
		}
	}
	projectQuery.Distinct("project_id").Count(&e.ProjectCount)

	yearFilterTask := ""
	yearFilterDone := ""
	if year != "" {
		if _, err := strconv.Atoi(year); err == nil {
			yearFilterTask = " AND EXTRACT(YEAR FROM t.created_at) = " + year
			yearFilterDone = " AND EXTRACT(YEAR FROM t.completed_at) = " + year
		}
	}

	type PriorityWeight struct {
		TotalWeight float64
	}
	var allWeight, doneWeight PriorityWeight
	db.PgSql.Raw(`
		SELECT COALESCE(SUM(tp.level), 0) as total_weight
		FROM tasks t
		JOIN task_priorities tp ON tp.no = t.priority_id
		WHERE t.user_id = ?`+yearFilterTask, userID).Scan(&allWeight)
	db.PgSql.Raw(`
		SELECT COALESCE(SUM(tp.level), 0) as total_weight
		FROM tasks t
		JOIN task_priorities tp ON tp.no = t.priority_id
		WHERE t.user_id = ? AND t.completed_at IS NOT NULL`+yearFilterDone, userID).Scan(&doneWeight)
	if allWeight.TotalWeight > 0 {
		e.TPS = doneWeight.TotalWeight / allWeight.TotalWeight * 100
	}

	type Efficiency struct {
		Value float64
	}
	var werResult Efficiency
	db.PgSql.Raw(`
		SELECT COALESCE(AVG((t.estimated_hours / NULLIF(t.actual_hours, 0)) * 100), 0) as value
		FROM tasks t
		WHERE t.user_id = ?
			AND t.completed_at IS NOT NULL
			AND t.estimated_hours > 0
			AND t.actual_hours > 0`+yearFilterDone, userID).Scan(&werResult)
	e.WER = werResult.Value

	e.FinalScore = e.TCR*0.3 + e.OTR*0.3 + e.TPS*0.2 + e.WER*0.2

	switch {
	case e.FinalScore >= 85:
		e.Category = "Sangat Baik"
	case e.FinalScore >= 70:
		e.Category = "Baik"
	case e.FinalScore >= 55:
		e.Category = "Cukup"
	case e.FinalScore >= 40:
		e.Category = "Buruk"
	default:
		e.Category = "Sangat Buruk"
	}

	db.PgSql.Raw(`
		SELECT ts.label, ts.color, COUNT(t.no) as count
		FROM tasks t
		JOIN task_statuses ts ON ts.no = t.status_id
		WHERE t.user_id = ? AND t.completed_at IS NOT NULL`+yearFilterDone+`
		GROUP BY ts.label, ts.color
		ORDER BY count DESC
	`, userID).Scan(&e.StatusDistribution)

	db.PgSql.Raw(`
		SELECT t.type, COUNT(t.no) as count
		FROM tasks t
		WHERE t.user_id = ? AND t.completed_at IS NOT NULL`+yearFilterDone+`
		GROUP BY t.type
		ORDER BY count DESC
	`, userID).Scan(&e.TypeDistribution)

	if year != "" {
		if _, err := strconv.Atoi(year); err == nil {
			db.PgSql.Raw(`
				SELECT EXTRACT(YEAR FROM completed_at)::int as year,
					   EXTRACT(MONTH FROM completed_at)::int as month,
					   COUNT(no) as count
				FROM tasks
				WHERE user_id = ?
					AND completed_at IS NOT NULL
					AND EXTRACT(YEAR FROM completed_at) = `+year+`
				GROUP BY year, month
				ORDER BY year, month
			`, userID).Scan(&e.MonthlyCompletion)
		}
	} else {
		db.PgSql.Raw(`
			SELECT EXTRACT(YEAR FROM completed_at)::int as year,
				   EXTRACT(MONTH FROM completed_at)::int as month,
				   COUNT(no) as count
			FROM tasks
			WHERE user_id = ?
				AND completed_at IS NOT NULL
				AND completed_at >= ?
			GROUP BY year, month
			ORDER BY year, month
		`, userID, time.Now().AddDate(0, -6, 0)).Scan(&e.MonthlyCompletion)
	}

	return e
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
