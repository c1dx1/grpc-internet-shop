package handlers

import (
	"context"
	"internet-shop/repository"
	"internet-shop/shared/proto"
)

type CartHandler struct {
	repo repository.CartRepository
	proto.UnimplementedCartServiceServer
}

func NewCartHandler(repo repository.CartRepository) *CartHandler {
	return &CartHandler{repo: repo}
}

func (h *CartHandler) GetCart(ctx context.Context, req *proto.AllFromCartRequest) (*proto.FullCartResponse, error) {
	cart, err := h.repo.GetCart(ctx, req.UserId)
	if err != nil {
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
	err := h.repo.AddToCart(ctx, req.UserId, req.ProductId, req.Quantity)
	if err != nil {
		return nil, err
	}

	return &proto.EmptyCartResponse{}, nil
}

func (h *CartHandler) RemoveFromCart(ctx context.Context, req *proto.CartRequest) (*proto.FullCartResponse, error) {
	err := h.repo.RemoveFromCart(ctx, req.UserId, req.ProductId)
	if err != nil {
		return nil, err
	}

	return h.GetCart(ctx, &proto.AllFromCartRequest{UserId: req.UserId})
}

func (h *CartHandler) UpdateCart(ctx context.Context, req *proto.CartRequest) (*proto.FullCartResponse, error) {
	err := h.repo.UpdateCart(ctx, req.UserId, req.ProductId, req.Quantity)
	if err != nil {
		return nil, err
	}

	return h.GetCart(ctx, &proto.AllFromCartRequest{UserId: req.UserId})
}
