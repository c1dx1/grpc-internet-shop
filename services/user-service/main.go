package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"internet-shop/repository"
	"internet-shop/services/user-service/handlers"
	"internet-shop/services/user-service/interceptors"
	"internet-shop/shared/config"
	"internet-shop/shared/messaging"
	"internet-shop/shared/proto"
	"log"
	"net"
)

const userQueue = "user_queue"

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
		log.Fatalf("cannot load config file: %v", err)
	}

	pool, err := NewDBPool(cfg.PostgresURL())
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
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

	_, err = messaging.DeclareQueue(rabbitCh, userQueue)
	if err != nil {
		log.Fatalf("Error declaring order_queue: %s", err)
	}

	redisClient, err := NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("cannot connect to redis: %v", err)
	}
	defer redisClient.Close()

	userRepo := repository.NewUserRepository(pool)
	userHandler := handlers.NewUserHandler(*userRepo, rabbitCh, redisClient)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptors.SessionAuthInterceptor(redisClient)))
	proto.RegisterUserServiceServer(grpcServer, userHandler)

	reflection.Register(grpcServer)

	listener, err := net.Listen(cfg.ServicesNetworkType, cfg.UserPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("User service started on port %s", cfg.UserPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
