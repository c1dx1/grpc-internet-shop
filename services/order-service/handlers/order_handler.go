package handlers

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"internet-shop/repository"
	"internet-shop/shared/models"
	"internet-shop/shared/proto"
	"log"
)

type OrderHandler struct {
	repo     repository.OrderRepository
	rabbitCh *amqp.Channel
	proto.UnimplementedOrderServiceServer
}

func NewOrderHandler(repo repository.OrderRepository, rabbitCh *amqp.Channel) *OrderHandler {
	return &OrderHandler{repo: repo, rabbitCh: rabbitCh}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.OrderIDResponse, error) {
	var products []models.Product
	var total float64

	for _, p := range req.Products {
		product := models.Product{
			ID:       p.Id,
			Name:     p.Name,
			Price:    p.Price,
			Quantity: p.Quantity,
		}
		products = append(products, product)
		total += p.Price * float64(p.Quantity)
	}

	order := models.Order{
		UserID:     req.UserId,
		Products:   products,
		TotalPrice: total,
	}

	orderId, err := h.repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	message := map[string]interface{}{
		"order_id": orderId,
		"products": products,
		"user_id":  req.UserId,
		"total":    total,
	}

	messageBody, _ := json.Marshal(message)

	err = h.rabbitCh.Publish(
		"",
		"order_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})

	if err != nil {
		log.Printf("Error publishing message: %s", err)
	}

	return &proto.OrderIDResponse{
		OrderId: orderId,
	}, nil
}

func (h *OrderHandler) GetOrderById(ctx context.Context, req *proto.OrderRequest) (*proto.OrderResponse, error) {
	order, err := h.repo.GetOrderById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	var products []*proto.Product
	for _, p := range order.Products {
		product := &proto.Product{
			Id:       p.ID,
			Name:     p.Name,
			Price:    p.Price,
			Quantity: p.Quantity,
		}
		products = append(products, product)
	}

	return &proto.OrderResponse{
		Id:         order.ID,
		UserId:     order.UserID,
		Product:    products,
		TotalPrice: order.TotalPrice,
	}, nil
}
