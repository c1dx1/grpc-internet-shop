package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"internet-shop/repository"
	"internet-shop/services/cart-service/handlers"
	"internet-shop/services/user-service/interceptors"
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

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
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

	redisClient, err := NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Error connecting to redis: %s", err)
	}

	productRepo := repository.NewProductRepository(pool)
	cartRepo := repository.NewCartRepository(pool, productRepo)
	cartHandler := handlers.NewCartHandler(*cartRepo)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptors.SessionAuthInterceptor(redisClient)))
	proto.RegisterCartServiceServer(grpcServer, cartHandler)

	reflection.Register(grpcServer)

	listener, err := net.Listen(cfg.ServicesNetworkType, cfg.CartPort)
	if err != nil {
		log.Fatalf("Error listening on port %s", cfg.CartPort)
	}
	log.Printf("Order service listening on port %s", cfg.CartPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error starting grpc server: %s", err)
	}
}
