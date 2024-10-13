package main

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"internet-shop/services/gateway-service/handlers"
	"internet-shop/shared/proto"
	"log"
)

func main() {
	router := gin.Default()

	productConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to Product service: %v", err)
	}
	defer productConn.Close()
	productClient := proto.NewProductServiceClient(productConn)

	orderConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to Order service: %v", err)
	}
	defer orderConn.Close()
	orderClient := proto.NewOrderServiceClient(orderConn)

	cartConn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to Cart service: %v", err)
	}
	defer cartConn.Close()
	cartClient := proto.NewCartServiceClient(cartConn)

	Handler := handlers.NewHandler(orderClient, productClient, cartClient)

	router.GET("/products", Handler.AllProductsHandler)
	router.GET("/products/:id", Handler.GetProductByIDHandler)
	router.POST("/orders", Handler.CreateOrderHandler)
	router.GET("/orders/:id", Handler.GetOrderByIdHandler)
	router.GET("/cart/:id", Handler.GetCartHandler)
	router.POST("/cart", Handler.AddToCartHandler)
	router.PUT("/cart", Handler.UpdateCartHandler)
	router.DELETE("/cart", Handler.RemoveFromCartHandler)

	log.Printf("Gateway service is listening on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("could not start gateway service: %v", err)
	}
}
