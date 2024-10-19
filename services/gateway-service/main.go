package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"internet-shop/services/gateway-service/handlers"
	"internet-shop/shared/config"
	"internet-shop/shared/proto"
	"log"
)

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
		log.Fatalf("config.LoadConfig failed: %v", err)
	}

	redisClient, err := NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("NewRedisClient failed: %v", err)
	}

	router := gin.Default()

	productConn, err := grpc.Dial(cfg.LocalhostURL(cfg.ProductPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to Product service: %v", err)
	}
	defer productConn.Close()
	productClient := proto.NewProductServiceClient(productConn)

	orderConn, err := grpc.Dial(cfg.LocalhostURL(cfg.OrderPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to Order service: %v", err)
	}
	defer orderConn.Close()
	orderClient := proto.NewOrderServiceClient(orderConn)

	cartConn, err := grpc.Dial(cfg.LocalhostURL(cfg.CartPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to Cart service: %v", err)
	}
	defer cartConn.Close()
	cartClient := proto.NewCartServiceClient(cartConn)

	userConn, err := grpc.Dial(cfg.LocalhostURL(cfg.UserPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to User service: %v", err)
	}
	defer userConn.Close()
	userClient := proto.NewUserServiceClient(userConn)

	Handler := handlers.NewHandler(orderClient, productClient, cartClient, userClient)

	router.GET("/products", Handler.AllProductsHandler)
	router.GET("/products/:id", Handler.GetProductByIDHandler)

	router.POST("/user/signup", Handler.SignUpHandler)
	router.POST("/user/signin", Handler.SignInHandler)

	authGroup := router.Group("/")
	authGroup.Use(sessionMiddleware(redisClient))
	{
		authGroup.POST("/orders", Handler.CreateOrderHandler)
		authGroup.GET("/orders/:id", Handler.GetOrderByIdHandler)

		router.GET("/cart", Handler.GetCartHandler)
		router.POST("/cart", Handler.AddToCartHandler)
		router.PUT("/cart", Handler.UpdateCartHandler)
		router.DELETE("/cart", Handler.RemoveFromCartHandler)

		authGroup.POST("/user/signout", Handler.SignOutHandler)
	}

	log.Printf("Gateway service is listening on port %s", cfg.GatewayPort)
	if err := router.Run(cfg.GatewayPort); err != nil {
		log.Fatalf("could not start gateway service: %v", err)
	}
}
