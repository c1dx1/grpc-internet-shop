package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"internet-shop/repository"
	"internet-shop/services/notification-service/handlers"
	email "internet-shop/services/notification-service/senders"
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

	redisClient, err := NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Error connecting to redis: %s", err)
	}

	eSend := email.NewEmailSender(cfg.SMTPFrom, cfg.SMTPUsername, cfg.SMTPPass, cfg.SMTPHost, cfg.SMTPPort)

	notificationRepo := repository.NewNotificationRepository(pool)
	notificationHandler := handlers.NewNotificationHandler(*notificationRepo)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptors.SessionAuthInterceptor(redisClient)))
	proto.RegisterNotificationServiceServer(grpcServer, notificationHandler)

	reflection.Register(grpcServer)

	go consumeNotificationMessage(*eSend, *notificationHandler, rabbitCh)

	listener, err := net.Listen(cfg.ServicesNetworkType, cfg.NotificationPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Notification service started on %s", cfg.NotificationPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
