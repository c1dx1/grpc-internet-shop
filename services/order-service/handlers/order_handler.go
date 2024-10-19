package handlers

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"internet-shop/repository"
	"internet-shop/shared/models"
	"internet-shop/shared/proto"
	"log"
	"strconv"
)

type OrderHandler struct {
	repo     repository.OrderRepository
	rabbitCh *amqp.Channel
	proto.UnimplementedOrderServiceServer
}

func NewOrderHandler(repo repository.OrderRepository, rabbitCh *amqp.Channel) *OrderHandler {
	return &OrderHandler{repo: repo, rabbitCh: rabbitCh}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
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

	userID, err := strconv.ParseInt(ctx.Value("user-id").(string), 10, 64)
	if err != nil {
		log.Printf("order_handler: createorder: strconv userid error: err:%v", err)
		return nil, err
	}

	order := models.Order{
		UserID:     userID,
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
		"user_id":  userID,
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

	return &proto.CreateOrderResponse{
		OrderId: orderId,
	}, nil
}

func (h *OrderHandler) GetOrderById(ctx context.Context, req *proto.OrderRequest) (*proto.OrderResponse, error) {
	userID, err := strconv.ParseInt(ctx.Value("user-id").(string), 10, 64)
	if err != nil {
		log.Printf("order_handler: getorder: strconv userid error: err:%v", err)
		return nil, err
	}

	order, err := h.repo.GetOrderById(ctx, userID, req.Id)
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
