package models

type Price struct {
	ProductID int     `json:"product_id"`
	CreatedAt string  `json:"created_at"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Price     float64 `json:"price"`
}