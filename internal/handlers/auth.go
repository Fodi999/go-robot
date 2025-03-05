package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"go-robot/internal/models"
)

// RegisterHandler – эндпоинт для регистрации гостей
func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
			return
		}
		var guest models.Guest
		if err := json.NewDecoder(r.Body).Decode(&guest); err != nil {
			http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
			return
		}
		if guest.Username == "" || guest.Email == "" || guest.Password == "" || guest.Phone == "" {
			http.Error(w, "Заполните все обязательные поля", http.StatusBadRequest)
			return
		}
		query := `INSERT INTO guests (username, email, password, phone) VALUES ($1, $2, $3, $4) RETURNING id`
		if err := db.QueryRow(query, guest.Username, guest.Email, guest.Password, guest.Phone).Scan(&guest.ID); err != nil {
			http.Error(w, "Ошибка сохранения в базе данных", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(guest)
	}
}

// LoginHandler – эндпоинт для входа
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
			return
		}
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
			return
		}
		if creds.Username == "" || creds.Password == "" {
			http.Error(w, "Заполните все обязательные поля", http.StatusBadRequest)
			return
		}
		var guest models.Guest
		query := `SELECT id, username, email, password, phone FROM guests WHERE username = $1 AND password = $2`
		if err := db.QueryRow(query, creds.Username, creds.Password).
			Scan(&guest.ID, &guest.Username, &guest.Email, &guest.Password, &guest.Phone); err != nil {
			http.Error(w, "Неверные учетные данные", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(guest)
	}
}