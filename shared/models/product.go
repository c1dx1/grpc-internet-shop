package models

type Product struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int32   `json:"quantity"`
}
