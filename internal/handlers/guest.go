package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"go-robot/internal/models"
)

// GuestHandler – эндпоинт для получения и обновления профиля гостя.
// URL: /guest/{id}
func GuestHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем id из URL
		idStr := r.URL.Path[len("/guest/"):]
		switch r.Method {
		case http.MethodGet:
			var guest models.Guest
			query := `SELECT id, username, email, password, phone FROM guests WHERE id = $1`
			if err := db.QueryRow(query, idStr).Scan(&guest.ID, &guest.Username, &guest.Email, &guest.Password, &guest.Phone); err != nil {
				http.Error(w, "Пользователь не найден", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(guest)
		case http.MethodPut:
			var guest models.Guest
			if err := json.NewDecoder(r.Body).Decode(&guest); err != nil {
				http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
				return
			}
			query := `UPDATE guests SET username = $1, email = $2, password = $3, phone = $4 WHERE id = $5`
			if _, err := db.Exec(query, guest.Username, guest.Email, guest.Password, guest.Phone, idStr); err != nil {
				http.Error(w, "Ошибка обновления профиля", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		}
	}
}

