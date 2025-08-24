package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var PgSql *gorm.DB

func PgSqlInit() {

	var err error
	dsn := os.Getenv("POSTGRES_URL")
	PgSql, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: true,
	})
	if err != nil {
		log.Fatal("Failed to connect to database!")
	}

}
