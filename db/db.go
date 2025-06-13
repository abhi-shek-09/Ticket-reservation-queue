package db

import (
	"log"
	"queue-system/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func InitPostgres(cfg *config.Config) *sqlx.DB {
	var err error
	DB, err = sqlx.Open("postgres", cfg.PostgresURL)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	err = DB.Ping()

	if err != nil {
		log.Fatal("Cannot ping DB: ", err)
	}

	log.Println("Connected to postgres")
	return DB
}

