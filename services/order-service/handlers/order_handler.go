package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"internet-shop/repository"
	"internet-shop/shared/config"
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

func NewProductConnection() (*grpc.ClientConn, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("order_handler: newpc: Error loading config: %s", err)
		return nil, err
	}
	conn, err := grpc.NewClient(fmt.Sprintf("localhost%s", cfg.ProductPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
		return nil, err
	}

	return conn, nil
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

	orderID, err := h.repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	err = h.UpdateQuantityMessage(products)
	if err != nil {
		log.Printf("order_handler: updateQuantity error: err:%v", err)
		return nil, err
	}

	err = h.SendNotification(userID, orderID)
	if err != nil {
		log.Printf("order_handler: sendnotif error: err:%v", err)
		return nil, err
	}

	return &proto.CreateOrderResponse{
		OrderId: orderID,
	}, nil
}

func (h *OrderHandler) GetOrderById(ctx context.Context, req *proto.OrderRequest) (*proto.OrderResponse, error) {
	userID, err := strconv.ParseInt(ctx.Value("user-id").(string), 10, 64)
	if err != nil {
		log.Printf("order_handler: getorder: strconv userid error: err:%v", err)
		return nil, err
	}

	productConn, err := NewProductConnection()
	defer productConn.Close()

	productClient := proto.NewProductServiceClient(productConn)

	if err != nil {
		log.Printf("order_handler: getorder: new product client error: err:%v", err)
		return nil, err
	}

	order, err := h.repo.GetOrderByID(ctx, userID, req.Id, productClient)
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

func (h *OrderHandler) UpdateQuantityMessage(products []models.Product) error {
	message := map[string]interface{}{
		"products": products,
	}

	messageBody, _ := json.Marshal(message)

	err := h.rabbitCh.Publish(
		"",
		"product_quantity_after_order",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})
	if err != nil {
		return fmt.Errorf("Error publishing message: %s", err)
	}

	return nil
}

func (h *OrderHandler) SendNotification(userID, orderID int64) error {
	message := map[string]interface{}{
		"user_id":  userID,
		"order_id": orderID,
		"author":   "OrderService",
		"subject":  "New Order!",
		"content":  fmt.Sprintf("Hello!\n\nYour order #%v was created successfuly.\n\nThank you!", orderID),
	}

	messageBody, _ := json.Marshal(message)

	err := h.rabbitCh.Publish(
		"",
		"notification_after_order",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})
	if err != nil {
		return fmt.Errorf("Error publishing message: %s", err)
	}

	return nil
}
