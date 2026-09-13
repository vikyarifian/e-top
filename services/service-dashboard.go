package services

import (
	"log/slog"
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
	TVS float64
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
	TCR         float64
	OTR         float64
	TVS         float64
	WER         float64
	FinalScore  float64
	Category    string
	ActiveRules []FuzzyRule
	// Evaluable bernilai false bila tidak ada tugas pada periode evaluasi.
	// Dalam keadaan itu seluruh indikator bernilai 0 dan inferensi akan
	// menghasilkan Sangat Buruk, padahal ketiadaan data bukan kinerja buruk.
	Evaluable bool

	TaskCount    int64
	DoneCount    int64
	OnTimeCount  int64
	ProjectCount int64
	// JudgedCount adalah cacah tugas yang nasib ketepatan waktunya sudah
	// pasti, yaitu tugas yang sudah selesai ditambah tugas yang belum selesai
	// padahal tenggatnya telah lewat. Angka ini menjadi penyebut OTR.
	JudgedCount int64

	StatusDistribution []StatusCount
	TypeDistribution   []TypeCount
	MonthlyCompletion  []MonthlyCount
}

// Task Value Score menimbang setiap tugas dengan dua dimensi yang keduanya
// tersimpan sebagai kolom weight pada tabel acuannya masing-masing:
//
//	task_priorities.weight  High 1,000  Medium 0,900  Low 0,800
//	task_impacts.weight     High 1,000  Medium 0,900  Low 0,800
//
// Nilai sebuah tugas adalah hasil kali kedua bobot itu, sehingga berkisar
// antara 0,640 dan 1,000. Tugas yang tidak memiliki acuan diabaikan dari
// perhitungan, bukan diberi bobot bawaan, agar tidak memihak.
const tvsWeightSQL = `COALESCE(tp.weight, 0) * COALESCE(ti.weight, 0)`

// sumTaskValue menjumlahkan nilai seluruh tugas yang cocok dengan syarat
// tambahan yang diberikan. Galat kueri dilaporkan, bukan dibiarkan menjadi
// nol diam-diam: bila migrasi 001_task_value_score.sql belum dijalankan pada
// basis data, tabel task_impacts tidak ada dan penjumlahan akan gagal. Tanpa
// pelaporan, kegagalan itu tampak sebagai TVS nol seolah-olah karyawan tidak
// menuntaskan satu tugas pun.
func sumTaskValue(where string, args ...any) float64 {
	var row struct{ TotalWeight float64 }
	q := `
		SELECT COALESCE(SUM(` + tvsWeightSQL + `), 0) as total_weight
		FROM tasks t
		JOIN task_priorities tp ON tp.no = t.priority_id
		JOIN task_impacts ti ON ti.no = t.impact_id
		WHERE ` + where
	if err := db.PgSql.Raw(q, args...).Scan(&row).Error; err != nil {
		slog.Error("perhitungan Task Value Score gagal",
			"error", err,
			"petunjuk", "jalankan db/migrations/001_task_value_score.sql pada basis data ini")
		return 0
	}
	return row.TotalWeight
}

func GetAchievedEvaluation(userID string, year string) AchievedEvaluation {
	var e AchievedEvaluation

	// Periode evaluasi ditentukan oleh tahun penugasan (created_at). .
	doneQuery := db.PgSql.Model(&models.Task{}).
		Where("user_id = ? AND completed_at IS NOT NULL", userID)
	if year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			doneQuery = doneQuery.Where("EXTRACT(YEAR FROM created_at) = ?", y)
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
			onTimeQuery = onTimeQuery.Where("EXTRACT(YEAR FROM created_at) = ?", y)
		}
	}
	onTimeQuery.Count(&e.OnTimeCount)

	// Penyebut OTR mencakup tugas yang sudah selesai dan tugas yang belum
	// selesai padahal tenggatnya sudah lewat. Yang kedua sudah pasti terlambat,
	// jadi tidak pantas dikecualikan. Tugas yang belum selesai dan tenggatnya
	// belum tiba masih mungkin tepat waktu, sehingga belum dinilai.
	judgedQuery := db.PgSql.Model(&models.Task{}).
		Where("user_id = ? AND (completed_at IS NOT NULL OR (due_date IS NOT NULL AND due_date < ?))",
			userID, time.Now())
	if year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			judgedQuery = judgedQuery.Where("EXTRACT(YEAR FROM created_at) = ?", y)
		}
	}
	judgedQuery.Count(&e.JudgedCount)

	if e.TaskCount > 0 {
		e.TCR = float64(e.DoneCount) / float64(e.TaskCount) * 100
	}
	if e.JudgedCount > 0 {
		e.OTR = float64(e.OnTimeCount) / float64(e.JudgedCount) * 100
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
			// kohor yang sama dipakai untuk agregasi tugas selesai
			yearFilterDone = yearFilterTask
		}
	}

	// Task Value Score: perbandingan nilai tugas yang tuntas terhadap nilai
	// seluruh tugas, dengan nilai tiap tugas diambil dari hasil kali bobot
	// prioritas dan bobot dampaknya.
	type PriorityWeight struct {
		TotalWeight float64
	}
	var allWeight, doneWeight PriorityWeight
	allWeight.TotalWeight = sumTaskValue("t.user_id = ?"+yearFilterTask, userID)
	doneWeight.TotalWeight = sumTaskValue("t.user_id = ? AND t.completed_at IS NOT NULL"+yearFilterDone, userID)
	if allWeight.TotalWeight > 0 {
		e.TVS = doneWeight.TotalWeight / allWeight.TotalWeight * 100
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
	if e.WER > 100 {
		e.WER = 100
	}

	e.Evaluable = e.TaskCount > 0
	if e.Evaluable {
		e.FinalScore, e.Category, e.ActiveRules = FuzzyTsukamoto(e.TCR, e.OTR, e.TVS, e.WER)
	} else {
		e.Category = "Belum Dapat Dinilai"
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

	// Pembilang dan penyebut harus memakai skala bobot yang sama, yaitu hasil
	// kali bobot prioritas dan bobot dampak.
	allWeight := sumTaskValue("t.user_id = ? OR t.created_by = ?", userID, userID)
	doneWeight := sumTaskValue("(t.user_id = ? OR t.created_by = ?) AND t.completed_at IS NOT NULL", userID, userID)
	if allWeight > 0 {
		d.TVS = doneWeight / allWeight * 100
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
	if d.WER > 100 {
		d.WER = 100
	}

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
