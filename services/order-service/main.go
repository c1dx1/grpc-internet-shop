package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"internet-shop/repository"
	"internet-shop/services/order-service/handlers"
	"internet-shop/services/user-service/interceptors"
	"internet-shop/shared/config"
	"internet-shop/shared/messaging"
	"internet-shop/shared/proto"
	"log"
	"net"
)

const productQuantityAfterOrderQueue = "product_quantity_after_order"
const notificationAfterOrderQueue = "notification_after_order"

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

	_, err = messaging.DeclareQueue(rabbitCh, productQuantityAfterOrderQueue)
	if err != nil {
		log.Fatalf("Error declaring product_quantity_after_order: %s", err)
	}

	_, err = messaging.DeclareQueue(rabbitCh, notificationAfterOrderQueue)
	if err != nil {
		log.Fatalf("Error declaring notification_after_order: %s", err)
	}

	redisClient, err := NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Error connecting to redis: %s", err)
	}

	orderRepo := repository.NewOrderRepository(pool)
	orderHandler := handlers.NewOrderHandler(*orderRepo, rabbitCh)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptors.SessionAuthInterceptor(redisClient)))
	proto.RegisterOrderServiceServer(grpcServer, orderHandler)

	reflection.Register(grpcServer)

	listener, err := net.Listen(cfg.ServicesNetworkType, cfg.OrderPort)
	if err != nil {
		log.Fatalf("Error listening on port %s", cfg.OrderPort)
	}
	log.Printf("Order service listening on port %s", cfg.OrderPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error starting grpc server: %s", err)
	}
}
