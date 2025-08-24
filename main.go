package main

import (
	"etop/db"
	"etop/routes"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load(".env.local"); err != nil {
			log.Fatal(err)
		}
	}

	db.PgSqlInit()

	routes.SetRoutes()
}
