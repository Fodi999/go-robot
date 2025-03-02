package db

import (
	"database/sql"
	"log"
	"os"
    "fmt"
	_ "github.com/lib/pq" // PostgreSQL драйвер
)

func Connect() (*sql.DB, error) {
	// Получаем DSN из переменной окружения
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("переменная окружения DATABASE_URL не установлена")
	}

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
