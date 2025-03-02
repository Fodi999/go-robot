package main

import (
	"log"
	"net/http"

	"go-robot/internal/chat"
	"go-robot/internal/db"
	"go-robot/internal/handlers"
	"go-robot/internal/seed"
)

func main() {
	// Подключаемся к базе данных
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer database.Close()

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

	// Запускаем сервер
	port := "8080"
	log.Printf("Сервер запущен на порту %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}







