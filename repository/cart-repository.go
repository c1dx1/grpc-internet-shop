package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"internet-shop/shared/models"
	"internet-shop/shared/proto"
)

type CartRepository struct {
	db *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetCart(ctx context.Context, userId int64, productService proto.ProductServiceClient) (models.Cart, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return models.Cart{}, err
	}
	defer conn.Release()

	var cart models.Cart

	err = conn.QueryRow(ctx, "SELECT user_id, total_price FROM carts WHERE user_id = $1", userId).
		Scan(&cart.UserID, &cart.TotalPrice)
	if err != nil {
		return models.Cart{}, err
	}

	rows, err := conn.Query(ctx, "SELECT product_id, quantity FROM cart_items WHERE user_id = $1", userId)
	if err != nil {
		return models.Cart{}, err
	}
	defer rows.Close()

	var products []models.Product

	for rows.Next() {
		var product models.Product

		err = rows.Scan(&product.ID, &product.Quantity)
		if err != nil {
			return models.Cart{}, err
		}

		pDB, err := productService.GetProductById(ctx, &proto.ProductRequest{Id: product.ID})
		if err != nil {
			return models.Cart{}, err
		}

		product.Name = pDB.Name
		product.Price = pDB.Price

		products = append(products, product)
	}

	cart.Products = products
	return cart, nil
}

func (r *CartRepository) AddToCart(ctx context.Context, userId int64, productId int64, quantity int32) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}

	defer conn.Release()

	_, err = conn.Exec(ctx, "INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)", userId, productId, quantity)
	if err != nil {
		return err
	}

	return nil
}

func (r *CartRepository) RemoveFromCart(ctx context.Context, userId int64, productId int64) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userId, productId)
	if err != nil {
		return err
	}

	return nil
}

func (r *CartRepository) UpdateCart(ctx context.Context, userId int64, productId int64, quantity int32) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "UPDATE cart_items SET quantity = $1 WHERE user_id = $2 AND product_id = $3", quantity, userId, productId)
	if err != nil {
		return err
	}
	return nil
}

func (r *CartRepository) UpdateTotalPrice(ctx context.Context, userId int64, productClient proto.ProductServiceClient) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	var totalPrice float64
	rows, err := conn.Query(ctx, "SELECT product_id, quantity FROM cart_items WHERE user_id = $1", userId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		err = rows.Scan(&product.ID, &product.Quantity)

		pDB, err := productClient.GetProductById(ctx, &proto.ProductRequest{Id: product.ID})
		if err != nil {
			return err
		}

		product.Price = pDB.Price

		totalPrice += product.Price * float64(product.Quantity)
	}

	_, err = conn.Exec(ctx, "UPDATE carts SET total_price = $1 WHERE user_id = $2", totalPrice, userId)
	if err != nil {
		return err
	}
	return nil
}

func (r *CartRepository) CreateCart(ctx context.Context, userID int64) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	var flag bool

	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM carts WHERE user_id=$1)", userID).Scan(&flag)
	if flag {
		return fmt.Errorf("Cart is alredy exists for this userid")
	}

	_, err = conn.Exec(ctx, "INSERT INTO carts (user_id, total_price) VALUES ($1, $2)", userID, 0)
	if err != nil {
		return err
	}

	return nil
}
