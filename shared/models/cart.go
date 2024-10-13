package models

type Cart struct {
	UserID     int64
	Products   []Product
	TotalPrice float64
}
