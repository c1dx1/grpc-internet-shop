package repository

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"internet-shop/shared/models"
)

type OrderRepository struct {
	db          *pgxpool.Pool
	productRepo *ProductRepository
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db, productRepo: NewProductRepository(db)}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order models.Order) (int64, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var orderID int64
	err = tx.QueryRow(ctx,
		"INSERT INTO orders (user_id, total_price) VALUES ($1, $2) RETURNING id",
		order.UserID, order.TotalPrice).
		Scan(&orderID)
	if err != nil {
		return 0, err
	}

	for _, product := range order.Products {
		_, err := tx.Exec(ctx,
			"INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)",
			orderID, product.ID, product.Quantity)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return orderID, nil
}

func (r *OrderRepository) GetOrderById(ctx context.Context, orderID int64) (models.Order, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return models.Order{}, err
	}
	defer conn.Release()

	var order models.Order

	err = conn.QueryRow(ctx, "SELECT id, user_id, total_price FROM orders WHERE id = $1", orderID).
		Scan(&order.ID, &order.UserID, &order.TotalPrice)
	if err != nil {
		return models.Order{}, err
	}

	rows, err := conn.Query(ctx, "SELECT product_id, quantity FROM order_items WHERE order_id = $1", orderID)
	if err != nil {
		return models.Order{}, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product

		err = rows.Scan(&p.ID, &p.Quantity)
		if err != nil {
			return models.Order{}, err
		}

		product, err := r.productRepo.GetProductById(ctx, p.ID)
		if err != nil {
			return models.Order{}, err
		}

		p.Name = product.Name
		p.Price = product.Price

		products = append(products, p)
	}
	order.Products = products

	return order, nil
}
