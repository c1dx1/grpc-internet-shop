package models

type Order struct {
	ID         int64
	UserID     int64
	Products   []Product
	TotalPrice float64
}
