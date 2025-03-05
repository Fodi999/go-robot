package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"go-robot/internal/models"
)

// ProductsHandler – эндпоинт для создания (POST) и получения (GET) списка продуктов.
func ProductsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var prod models.Product
			if err := json.NewDecoder(r.Body).Decode(&prod); err != nil {
				http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
				return
			}
			if prod.Title == "" || prod.Description == "" || prod.Price == "" || prod.Calories == 0 ||
				prod.Category == "" || prod.ImageURL == "" {
				http.Error(w, "Заполните все обязательные поля", http.StatusBadRequest)
				return
			}
			insertQuery := `
			INSERT INTO products (title, description, price, calories, category, image_url)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, title, description, price, calories, category, image_url`
			if err := db.QueryRow(insertQuery, prod.Title, prod.Description, prod.Price, prod.Calories, prod.Category, prod.ImageURL).
				Scan(&prod.ID, &prod.Title, &prod.Description, &prod.Price, &prod.Calories, &prod.Category, &prod.ImageURL); err != nil {
				http.Error(w, "Ошибка сохранения продукта", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(prod)
		case http.MethodGet:
			rows, err := db.Query(`SELECT id, title, description, price, calories, category, image_url FROM products`)
			if err != nil {
				http.Error(w, "Ошибка получения продуктов", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			products := []models.Product{}
			for rows.Next() {
				var p models.Product
				if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.Calories, &p.Category, &p.ImageURL); err != nil {
					http.Error(w, "Ошибка сканирования продукта", http.StatusInternalServerError)
					return
				}
				products = append(products, p)
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(products)
		default:
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		}
	}
}

// ProductUpdateHandler – эндпоинт для редактирования продукта по ID (PUT)
// URL должен иметь вид: /products/{id}
func ProductUpdateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем id из URL
		idStr := r.URL.Path[len("/products/"):]
		if r.Method != http.MethodPut {
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
			return
		}
		var prod models.Product
		if err := json.NewDecoder(r.Body).Decode(&prod); err != nil {
			http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
			return
		}
		if prod.Title == "" || prod.Description == "" || prod.Price == "" || prod.Calories == 0 ||
			prod.Category == "" || prod.ImageURL == "" {
			http.Error(w, "Заполните все обязательные поля", http.StatusBadRequest)
			return
		}
		updateQuery := `
		UPDATE products
		SET title = $1, description = $2, price = $3, calories = $4, category = $5, image_url = $6
		WHERE id = $7
		RETURNING id, title, description, price, calories, category, image_url`
		if err := db.QueryRow(updateQuery, prod.Title, prod.Description, prod.Price, prod.Calories, prod.Category, prod.ImageURL, idStr).
			Scan(&prod.ID, &prod.Title, &prod.Description, &prod.Price, &prod.Calories, &prod.Category, &prod.ImageURL); err != nil {
			http.Error(w, "Ошибка обновления продукта", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(prod)
	}
}