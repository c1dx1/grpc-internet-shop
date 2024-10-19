package handlers

import (
	"context"
	"internet-shop/repository"
	"internet-shop/shared/proto"
	"log"
	"strconv"
)

type CartHandler struct {
	repo repository.CartRepository
	proto.UnimplementedCartServiceServer
}

func NewCartHandler(repo repository.CartRepository) *CartHandler {
	return &CartHandler{repo: repo}
}

func (h *CartHandler) GetCart(ctx context.Context, req *proto.EmptyCartRequest) (*proto.FullCartResponse, error) {
	userID, err := strconv.ParseInt(ctx.Value("user-id").(string), 10, 64)
	if err != nil {
		log.Printf("cart_handler: GetCart: strconv.ParseInt err: %v", err)
		return nil, err
	}

	cart, err := h.repo.GetCart(ctx, userID)
	if err != nil {
		log.Printf("cart_handler: GetCart repo.GetCart err:%v", err)
		return nil, err
	}

	var products []*proto.Product
	for _, p := range cart.Products {
		products = append(products, &proto.Product{
			Id:       p.ID,
			Name:     p.Name,
			Price:    p.Price,
			Quantity: p.Quantity,
		})
	}

	return &proto.FullCartResponse{
		UserId:     cart.UserID,
		Product:    products,
		TotalPrice: cart.TotalPrice,
	}, nil
}

func (h *CartHandler) AddToCart(ctx context.Context, req *proto.CartRequest) (*proto.EmptyCartResponse, error) {
	userID, err := strconv.ParseInt(ctx.Value("user-id").(string), 10, 64)
	if err != nil {
		log.Printf("cart_handler: AddToCart: strconv.ParseInt err: %v", err)
		return nil, err
	}
	err = h.repo.AddToCart(ctx, userID, req.ProductId, req.Quantity)
	if err != nil {
		log.Printf("cart_handler: AddToCart: repo.AddToCart err: %v", err)
		return nil, err
	}

	return &proto.EmptyCartResponse{}, nil
}

func (h *CartHandler) RemoveFromCart(ctx context.Context, req *proto.CartRequest) (*proto.FullCartResponse, error) {
	userID, err := strconv.ParseInt(ctx.Value("user-id").(string), 10, 64)
	if err != nil {
		log.Printf("cart_handler: RemoveFromCart: strconv.ParseInt err: %v", err)
		return nil, err
	}

	err = h.repo.RemoveFromCart(ctx, userID, req.ProductId)
	if err != nil {
		log.Printf("cart_handler: RemoveFromCart: repo.RemoveFromCart err: %v", err)
		return nil, err
	}

	return h.GetCart(ctx, &proto.EmptyCartRequest{})
}

func (h *CartHandler) UpdateCart(ctx context.Context, req *proto.CartRequest) (*proto.FullCartResponse, error) {
	userID, err := strconv.ParseInt(ctx.Value("user-id").(string), 10, 64)
	if err != nil {
		log.Printf("cart_handler: UpdateCart: strconv.ParseInt err: %v", err)
		return nil, err
	}

	err = h.repo.UpdateCart(ctx, userID, req.ProductId, req.Quantity)
	if err != nil {
		log.Printf("cart_handler: UpdateCart: repo.UpdateCart err: %v", err)
		return nil, err
	}

	return h.GetCart(ctx, &proto.EmptyCartRequest{})
}
