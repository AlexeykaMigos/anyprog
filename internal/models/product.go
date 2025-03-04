package models

type Product struct {
	ID          int     `db:"id" json:"id"`
	Title       string  `db:"title" json:"title"`
	Description *string `db:"description" json:"description"`
	Price       float64 `db:"price" json:"price"`
	Version     int     `db:"version" json:"version"`
}
