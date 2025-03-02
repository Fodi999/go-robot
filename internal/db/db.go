package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // PostgreSQL драйвер
)

func Connect() (*sql.DB, error) {
	// DSN для подключения к базе данных Neon
	dsn := "postgresql://neondb_owner:npg_6tbnCaKfT5uZ@ep-floral-voice-a9hfvitj-pooler.gwc.azure.neon.tech/neondb?sslmode=require"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Успешное подключение к базе данных Neon!")
	return db, nil
}
