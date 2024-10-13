package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"internet-shop/repository"
	"internet-shop/services/order-service/handlers"
	"internet-shop/shared/config"
	"internet-shop/shared/messaging"
	"internet-shop/shared/proto"
	"log"
	"net"
)

const orderQueue = "order_queue"

func NewDBPool(connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	pool, err := NewDBPool(cfg.PostgresURL())
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
	defer pool.Close()

	rabbitConn, err := messaging.NewRabbitMQConnections(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Error connecting to rabbitmq: %s", err)
	}
	defer rabbitConn.Close()

	rabbitCh, err := messaging.NewRabbitMQChannels(rabbitConn)
	if err != nil {
		log.Fatalf("Error creating rabbitmq channel: %s", err)
	}
	defer rabbitCh.Close()

	_, err = messaging.DeclareQueue(rabbitCh, orderQueue)
	if err != nil {
		log.Fatalf("Error declaring order_queue: %s", err)
	}

	orderRepo := repository.NewOrderRepository(pool)
	orderHandler := handlers.NewOrderHandler(*orderRepo, rabbitCh)

	grpcServer := grpc.NewServer()
	proto.RegisterOrderServiceServer(grpcServer, orderHandler)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Error listening on port 50052")
	}
	log.Println("Order service listening on port 50052")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error starting grpc server: %s", err)
	}
}
