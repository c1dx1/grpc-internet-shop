package repository

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"internet-shop/shared/models"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT id, name, price, quantity FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) GetProductById(ctx context.Context, id int64) (models.Product, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return models.Product{}, err
	}
	defer conn.Release()

	var product models.Product
	err = conn.QueryRow(ctx, "SELECT id, name, price, quantity FROM products WHERE id = $1", id).Scan(&product.ID, &product.Name, &product.Price, &product.Quantity)
	if err != nil {
		return models.Product{}, err
	}

	return product, nil
}

func (r *ProductRepository) UpdateProductQuantity(productID int64, quantity int32) error {
	conn, err := r.db.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	var currentQuantity int32
	err = conn.QueryRow(context.Background(), "SELECT quantity FROM products WHERE id = $1", productID).Scan(&currentQuantity)
	if err != nil {
		return err
	}

	_, err = conn.Exec(context.Background(), "UPDATE products SET quantity = $1 WHERE id = $2", currentQuantity-quantity, productID)
	if err != nil {
		return err
	}

	return nil
}
