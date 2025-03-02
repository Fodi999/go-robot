package seed

import (
	"database/sql"
	"log"
)

// InsertSampleProducts добавляет или обновляет карточки продуктов в базе,
// а затем удаляет дубликаты (оставляя только уникальные записи).
func InsertSampleProducts(db *sql.DB) {
	// Используем карточки из файла sample_products.go (SampleProducts)
	for _, p := range SampleProducts {
		var existingID int
		// Проверяем наличие продукта с одинаковым названием и ссылкой на изображение.
		err := db.QueryRow("SELECT id FROM products WHERE title = $1 AND image_url = $2", p.Title, p.ImageURL).Scan(&existingID)
		if err == nil {
			// Если продукт существует, обновляем его данные.
			updateQuery := `
				UPDATE products
				SET description = $1, price = $2, calories = $3, category = $4
				WHERE id = $5`
			_, err = db.Exec(updateQuery, p.Description, p.Price, p.Calories, p.Category, existingID)
			if err != nil {
				log.Printf("Ошибка обновления продукта %s (id=%d): %v", p.Title, existingID, err)
			} else {
				log.Printf("Продукт %s (id=%d) успешно обновлён", p.Title, existingID)
			}
			continue
		}
		if err != sql.ErrNoRows {
			log.Printf("Ошибка проверки продукта %s: %v", p.Title, err)
			continue
		}

		// Если продукта нет, выполняем вставку.
		insertQuery := `
			INSERT INTO products (title, description, price, calories, category, image_url)
			VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = db.Exec(insertQuery, p.Title, p.Description, p.Price, p.Calories, p.Category, p.ImageURL)
		if err != nil {
			log.Printf("Ошибка вставки продукта %s: %v", p.Title, err)
		} else {
			log.Printf("Продукт %s успешно добавлен", p.Title)
		}
	}

	// Удаляем дубликаты. Оставляем запись с минимальным id для каждой уникальной группы,
	// определяемой по комбинации title, description и image_url.
	deleteQuery := `
		DELETE FROM products
		WHERE id NOT IN (
			SELECT MIN(id)
			FROM products
			GROUP BY title, description, image_url
		)`
	_, err := db.Exec(deleteQuery)
	if err != nil {
		log.Printf("Ошибка удаления дубликатов: %v", err)
	} else {
		log.Println("Дубликаты успешно удалены, остались только уникальные карточки продуктов")
	}
}




