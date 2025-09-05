package main

import (
	"etop/db"
	"etop/handlers"
	"etop/models"
	"etop/routes"
	"log"
	"log/slog"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Set default location jadi Asia/Jakarta
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}
	time.Local = loc // override global default
	slog.Info("Started", "Time", time.Now())
	slog.Info("Local", "Zone", time.Now().Location())
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load(".env.local"); err != nil {
			log.Fatal(err)
		}
	}

	handlers.OAthConfig()
	db.PgSqlInit()

	ps := []models.ProjectStatus{}
	err = db.PgSql.Find(&ps).Error

	if len(ps) == 0 || err != nil {
		ps = append(ps, models.ProjectStatus{Status: "PLANNING", Label: "Planning"})
		ps = append(ps, models.ProjectStatus{Status: "IN_PROGRESS", Label: "In Progress"})
		ps = append(ps, models.ProjectStatus{Status: "ON_HOLD", Label: "On Hold"})
		ps = append(ps, models.ProjectStatus{Status: "COMPLETED", Label: "Completed"})
		ps = append(ps, models.ProjectStatus{Status: "CANCELLED", Label: "Cancelled"})
		if err := db.PgSql.Create(&ps).Error; err != nil {
			println(err.Error())
		}
	}

	routes.SetRoutes()
}
