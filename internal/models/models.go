package models

import "time"

// Guest – структура пользователя
type Guest struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

// Product – структура товара (карточки)
type Product struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Calories    int    `json:"calories"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url"`
}

// Order – структура заказа
type Order struct {
	ID            int       `json:"id"`
	GuestID       int       `json:"guest_id"`
	ProductIDs    []int     `json:"product_ids"` // список идентификаторов товаров
	TotalPrice    string    `json:"total_price"`
	TotalCalories int       `json:"total_calories"`
	CreatedAt     time.Time `json:"created_at"`
}
