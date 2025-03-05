package main

import (
	"log"
	"net/http"
	"os" // Для работы с переменными окружения

	"go-robot/internal/chat"
	"go-robot/internal/db"
	"go-robot/internal/handlers"
	"go-robot/internal/seed"
	"github.com/joho/godotenv" // Для локальной разработки с .env
)

func main() {
	// Загружаем .env для локальной разработки (опционально)
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения")
	}

	// Логируем DATABASE_URL перед подключением
	log.Printf("DATABASE_URL: %s", os.Getenv("DATABASE_URL"))

	// Подключаемся к базе данных
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer database.Close()

	// Подсчитываем количество записей в таблице guests при запуске
	var count int
	countQuery := `SELECT COUNT(*) FROM guests`
	if err := database.QueryRow(countQuery).Scan(&count); err != nil {
		log.Printf("Ошибка при подсчёте данных при запуске: %v", err)
	} else {
		log.Printf("Количество сохранённых данных при запуске: %d", count)
	}

	// Заполняем таблицу продуктов начальными данными (если требуется)
	seed.InsertSampleProducts(database)

	// Инициализируем чат-хаб для WebSocket
	hub := chat.NewChatHub()
	go hub.Run() // Запускаем обработку сообщений чата в отдельной горутине

	// Регистрируем обработчики API
	http.HandleFunc("/register", handlers.RegisterHandler(database))
	http.HandleFunc("/login", handlers.LoginHandler(database))
	http.HandleFunc("/guest/", handlers.GuestHandler(database))
	http.HandleFunc("/products", handlers.ProductsHandler(database))
	http.HandleFunc("/products/", handlers.ProductUpdateHandler(database))
	http.HandleFunc("/orders", handlers.OrdersHandler(database))
	http.HandleFunc("/health", handlers.HealthHandler)
	// Регистрируем новый API-эндпоинт для общего количества клиентов
	http.HandleFunc("/api/total-customers", handlers.TotalCustomersHandler(database))
	// Подключаем WebSocket-обработчик
	http.HandleFunc("/ws", hub.ChatHandler)

	// Включаем CORS для всех маршрутов (если требуется)
	handler := handlers.EnableCORS(http.DefaultServeMux)

	// Получаем порт из переменной окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Значение по умолчанию
	}

	// Получаем WebSocket URL и логируем его
	wsURL := os.Getenv("WS_URL")
	if wsURL == "" {
		wsURL = "wss://localhost:" + port + "/ws" // Значение по умолчанию для локальной разработки
	}
	log.Printf("WebSocket URL: %s", wsURL)

	// Запускаем сервер
	log.Printf("Сервер запущен на порту %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}






