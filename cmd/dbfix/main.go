package main

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := dsnDariEnv()
	if dsn == "" {
		fmt.Println("DSN basis data belum diatur. Setel env SEED_DSN atau POSTGRES_URL.")
		return
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Println("connect error:", err)
		return
	}

	cmd := "check"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "check":
		check(db)
	case "dedup":
		dedup(db)
	}
}

func check(db *gorm.DB) {
	type S struct {
		Desc string
		Cnt  int64
	}
	var rows []S
	db.Raw(`SELECT COALESCE(details->>'description','<none>') as desc, COUNT(*) as cnt
		FROM logs WHERE action='completed_task' GROUP BY 1 ORDER BY cnt DESC LIMIT 5`).Scan(&rows)
	fmt.Println("=== completed_task descriptions ===")
	for _, r := range rows {
		fmt.Printf("  %-60q %d\n", r.Desc, r.Cnt)
	}
	var mismatch int64
	db.Raw(`SELECT COUNT(*) FROM logs l JOIN tasks t ON t.id = l.resource_id
		WHERE l.action='completed_task' AND l.created_at IS DISTINCT FROM t.completed_at`).Scan(&mismatch)
	fmt.Println("completed_task ts mismatch vs tasks.completed_at:", mismatch)
}

func dedup(db *gorm.DB) {
	res := db.Exec(`DELETE FROM logs
		WHERE action = 'updated_task'
		  AND details->>'description' = 'updated task status to Done and marked it as completed'`)
	fmt.Printf("duplicate updated_task completion logs deleted: %d (err=%v)\n", res.RowsAffected, res.Error)

	res = db.Exec(`UPDATE logs l
		SET created_at = t.completed_at, updated_at = t.completed_at
		FROM tasks t
		WHERE t.id = l.resource_id
		  AND l.action = 'completed_task'
		  AND t.completed_at IS NOT NULL
		  AND l.created_at IS DISTINCT FROM t.completed_at`)
	fmt.Printf("completed_task timestamps resynced: %d (err=%v)\n", res.RowsAffected, res.Error)

	type A struct {
		Action string
		Cnt    int64
	}
	var acts []A
	db.Raw(`SELECT action, COUNT(*) as cnt FROM logs GROUP BY action ORDER BY cnt DESC`).Scan(&acts)
	fmt.Println("=== actions after ===")
	for _, a := range acts {
		fmt.Printf("  %-20s %d\n", a.Action, a.Cnt)
	}
}

// dsnDariEnv membaca DSN basis data dari lingkungan. Kredensial sengaja tidak
// ditanam di dalam kode agar tidak ikut terpublikasi bersama repositori.
func dsnDariEnv() string {
	if v := os.Getenv("SEED_DSN"); v != "" {
		return v
	}
	return os.Getenv("POSTGRES_URL")
}
