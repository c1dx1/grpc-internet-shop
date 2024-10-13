package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"internet-shop/repository"
	"internet-shop/services/cart-service/handlers"
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

	productRepo := repository.NewProductRepository(pool)
	cartRepo := repository.NewCartRepository(pool, productRepo)
	cartHandler := handlers.NewCartHandler(*cartRepo)

	grpcServer := grpc.NewServer()
	proto.RegisterCartServiceServer(grpcServer, cartHandler)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Error listening on port 50053")
	}
	log.Println("Order service listening on port 50053")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error starting grpc server: %s", err)
	}
}
