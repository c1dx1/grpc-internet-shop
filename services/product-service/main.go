package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"internet-shop/repository"
	"internet-shop/services/product-service/handlers"
	"internet-shop/shared/config"
	"internet-shop/shared/proto"
	"log"
	"net"
)

func NewDBPool(connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func NewRabbitMQChannel(rabbitURL string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := NewDBPool(cfg.PostgresURL())
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
	defer pool.Close()

	rabbitConn, rabbitCh, err := NewRabbitMQChannel(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %s", err)
	}
	defer rabbitConn.Close()
	defer rabbitCh.Close()

	productRepo := repository.NewProductRepository(pool)
	productHandler := handlers.NewProductHandler(*productRepo)

	grpcServer := grpc.NewServer()
	proto.RegisterProductServiceServer(grpcServer, productHandler)

	reflection.Register(grpcServer)

	go consumeOrderMessage(*productRepo, rabbitCh)

	listener, err := net.Listen(cfg.ServicesNetworkType, cfg.ProductPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Product service started on %s", cfg.ProductPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
