package main

import (
	"log"
	"log/slog"
	"time"

	"github.com/joho/godotenv"

	"etop/db"
	"etop/handlers"
	"etop/routes"
	"etop/utils"
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

	utils.DbMutation()

	routes.SetRoutes()
}
