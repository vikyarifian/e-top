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

	// err := utils.SendVerificationEmail("vikyarifiansyah@gmail.com", "https://example.com/verify?token=abc123")
	// if err != nil {
	// 	log.Fatal("failed to send verification email:", err)
	// }

	db.PgSqlInit()

	routes.SetRoutes()
}
