package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL драйвер
)

// Connect подключается к базе данных и возвращает объект *sql.DB
func Connect() (*sql.DB, error) {
	// Получаем строку подключения из переменной окружения
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("переменная окружения DATABASE_URL не установлена")
	}

	// Открываем соединение с базой данных
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Успешное подключение к базе данных!")

	// Подсчёт общего количества записей
	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM guests").Scan(&totalCount)
	if err != nil {
		log.Printf("Ошибка при подсчёте общего количества записей: %v", err)
	} else {
		log.Printf("Общее количество записей в таблице guests: %d", totalCount)
	}

	// Подсчёт уникальных Username
	var uniqueUsernameCount int
	err = db.QueryRow("SELECT COUNT(DISTINCT username) FROM guests").Scan(&uniqueUsernameCount)
	if err != nil {
		log.Printf("Ошибка при подсчёте уникальных Username: %v", err)
	} else {
		log.Printf("Количество уникальных Username: %d", uniqueUsernameCount)
	}

	// Подсчёт уникальных Email
	var uniqueEmailCount int
	err = db.QueryRow("SELECT COUNT(DISTINCT email) FROM guests").Scan(&uniqueEmailCount)
	if err != nil {
		log.Printf("Ошибка при подсчёте уникальных Email: %v", err)
	} else {
		log.Printf("Количество уникальных Email: %d", uniqueEmailCount)
	}

	// Подсчёт уникальных Password
	var uniquePasswordCount int
	err = db.QueryRow("SELECT COUNT(DISTINCT password) FROM guests").Scan(&uniquePasswordCount)
	if err != nil {
		log.Printf("Ошибка при подсчёте уникальных Password: %v", err)
	} else {
		log.Printf("Количество уникальных Password: %d", uniquePasswordCount)
	}

	// Подсчёт уникальных Phone
	var uniquePhoneCount int
	err = db.QueryRow("SELECT COUNT(DISTINCT phone) FROM guests").Scan(&uniquePhoneCount)
	if err != nil {
		log.Printf("Ошибка при подсчёте уникальных Phone: %v", err)
	} else {
		log.Printf("Количество уникальных Phone: %d", uniquePhoneCount)
	}

	// Подсчёт уникальных FirstName
	var uniqueFirstNameCount int
	err = db.QueryRow("SELECT COUNT(DISTINCT first_name) FROM guests").Scan(&uniqueFirstNameCount)
	if err != nil {
		log.Printf("Ошибка при подсчёте уникальных FirstName: %v", err)
	} else {
		log.Printf("Количество уникальных FirstName: %d", uniqueFirstNameCount)
	}

	// Подсчёт уникальных Address
	var uniqueAddressCount int
	err = db.QueryRow("SELECT COUNT(DISTINCT address) FROM guests").Scan(&uniqueAddressCount)
	if err != nil {
		log.Printf("Ошибка при подсчёте уникальных Address: %v", err)
	} else {
		log.Printf("Количество уникальных Address: %d", uniqueAddressCount)
	}

	// Подсчёт уникальных BirthDate
	var uniqueBirthDateCount int
	err = db.QueryRow("SELECT COUNT(DISTINCT birth_date) FROM guests").Scan(&uniqueBirthDateCount)
	if err != nil {
		log.Printf("Ошибка при подсчёте уникальных BirthDate: %v", err)
	} else {
		log.Printf("Количество уникальных BirthDate: %d", uniqueBirthDateCount)
	}

	return db, nil
}