package main

import (
	"etop/db"
	"etop/handlers"
	"etop/routes"
	"fmt"
	"log"
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
	fmt.Println("time.Now():", time.Now())
	fmt.Println("Local zone:", time.Now().Location())
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load(".env.local"); err != nil {
			log.Fatal(err)
		}
	}

	handlers.OAthConfig()
	db.PgSqlInit()

	routes.SetRoutes()
}
