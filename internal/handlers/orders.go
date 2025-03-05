package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"go-robot/internal/models"
)

// OrdersHandler – эндпоинт для оформления заказа (POST) и получения истории заказов (GET)
func OrdersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req struct {
				GuestID    int   `json:"guest_id"`
				ProductIDs []int `json:"product_ids"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
				return
			}

			var totalCalories int
			var totalPrice float64 // Итоговая сумма в числовом формате.
			// Реальная логика расчёта: пробегаем по каждому product_id и суммируем цены и калории.
			for _, pid := range req.ProductIDs {
				var priceStr string
				var calories int
				if err := db.QueryRow(`SELECT price, calories FROM products WHERE id = $1`, pid).Scan(&priceStr, &calories); err != nil {
					http.Error(w, "Ошибка получения данных о продукте", http.StatusInternalServerError)
					return
				}
				var price float64
				// Предположим, что цена хранится в виде "$10". Убираем долларовый знак.
				_, err := fmt.Sscanf(priceStr, "$%f", &price)
				if err != nil {
					http.Error(w, "Ошибка обработки цены продукта", http.StatusInternalServerError)
					return
				}
				totalCalories += calories
				totalPrice += price
			}

			// Преобразуем список product_ids в JSON.
			productIDsJSON, err := json.Marshal(req.ProductIDs)
			if err != nil {
				http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
				return
			}

			var order models.Order
			insertOrderQuery := `
			INSERT INTO orders (guest_id, product_ids, total_price, total_calories)
			VALUES ($1, $2, $3, $4)
			RETURNING id, guest_id, product_ids, total_price, total_calories, created_at`
			if err := db.QueryRow(insertOrderQuery, req.GuestID, productIDsJSON, totalPrice, totalCalories).
				Scan(&order.ID, &order.GuestID, &order.ProductIDs, &order.TotalPrice, &order.TotalCalories, &order.CreatedAt); err != nil {
				http.Error(w, "Ошибка сохранения заказа", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(order)
		case http.MethodGet:
			guestIDStr := r.URL.Query().Get("guest_id")
			if guestIDStr == "" {
				http.Error(w, "guest_id не указан", http.StatusBadRequest)
				return
			}
			rows, err := db.Query(`SELECT id, guest_id, product_ids, total_price, total_calories, created_at FROM orders WHERE guest_id = $1`, guestIDStr)
			if err != nil {
				http.Error(w, "Ошибка получения заказов", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			orders := []models.Order{}
			for rows.Next() {
				var o models.Order
				var productIDsJSON []byte
				if err := rows.Scan(&o.ID, &o.GuestID, &productIDsJSON, &o.TotalPrice, &o.TotalCalories, &o.CreatedAt); err != nil {
					http.Error(w, "Ошибка сканирования заказа", http.StatusInternalServerError)
					return
				}
				if err := json.Unmarshal(productIDsJSON, &o.ProductIDs); err != nil {
					http.Error(w, "Ошибка обработки данных заказа", http.StatusInternalServerError)
					return
				}
				orders = append(orders, o)
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(orders)
		default:
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		}
	}
}

