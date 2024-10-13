package handlers

import (
	"context"
	"internet-shop/repository"
	"internet-shop/shared/proto"
	"log"
)

type ProductHandler struct {
	repo repository.ProductRepository
	proto.UnimplementedProductServiceServer
}

func NewProductHandler(repo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (h *ProductHandler) GetAllProducts(ctx context.Context, _ *proto.Empty) (*proto.ProductListResponse, error) {
	products, err := h.repo.GetAllProducts(ctx)
	if err != nil {
		log.Printf("Error getting products: %v", err)
		return nil, err
	}

	var productProducts []*proto.Product
	for _, product := range products {
		productProducts = append(productProducts, &proto.Product{
			Id:       product.ID,
			Name:     product.Name,
			Price:    product.Price,
			Quantity: product.Quantity,
		})
	}

	return &proto.ProductListResponse{Products: productProducts}, nil
}

func (h *ProductHandler) GetProductById(ctx context.Context, req *proto.ProductRequest) (*proto.ProductResponse, error) {
	product, err := h.repo.GetProductById(ctx, req.Id)
	if err != nil {
		log.Printf("Error getting product: %v", err)
		return nil, err
	}

	return &proto.ProductResponse{
		Id:       product.ID,
		Name:     product.Name,
		Price:    product.Price,
		Quantity: product.Quantity,
	}, nil
}

func (h *ProductHandler) UpdateProductQuantity(productID int64, quantity int32) error {
	err := h.repo.UpdateProductQuantity(productID, quantity)
	if err != nil {
		log.Printf("Error updating product quantity: %v", err)
		return err
	}
	return nil
}
