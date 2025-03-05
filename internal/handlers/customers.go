package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// TotalCustomersHandler возвращает количество зарегистрированных клиентов.
func TotalCustomersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var total int
		query := `SELECT COUNT(*) FROM guests`
		if err := db.QueryRow(query).Scan(&total); err != nil {
			http.Error(w, "Ошибка получения данных", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]int{"total": total})
	}
}