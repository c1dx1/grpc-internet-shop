package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"internet-shop/shared/proto"
	"net/http"
	"strconv"
)

type Handlers struct {
	productClient proto.ProductServiceClient
	orderClient   proto.OrderServiceClient
	cartClient    proto.CartServiceClient
}

func NewHandler(orderClient proto.OrderServiceClient, productClient proto.ProductServiceClient, cartClient proto.CartServiceClient) *Handlers {
	return &Handlers{
		orderClient:   orderClient,
		productClient: productClient,
		cartClient:    cartClient,
	}
}

func (h *Handlers) AllProductsHandler(c *gin.Context) {
	products, err := h.productClient.GetAllProducts(context.Background(), &proto.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *Handlers) GetProductByIDHandler(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productClient.GetProductById(context.Background(), &proto.ProductRequest{Id: productID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handlers) CreateOrderHandler(c *gin.Context) {
	var createOrderRequest proto.CreateOrderRequest
	if err := c.ShouldBindJSON(&createOrderRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderClient.CreateOrder(context.Background(), &createOrderRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handlers) GetOrderByIdHandler(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderClient.GetOrderById(context.Background(), &proto.OrderRequest{Id: orderID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handlers) GetCartHandler(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := h.cartClient.GetCart(context.Background(), &proto.AllFromCartRequest{UserId: userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, cart)
}

func (h *Handlers) AddToCartHandler(c *gin.Context) {
	var cartRequest proto.CartRequest
	if err := c.ShouldBindJSON(&cartRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.cartClient.AddToCart(context.Background(), &cartRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product added to cart"})
}

func (h *Handlers) RemoveFromCartHandler(c *gin.Context) {
	var cartRequest proto.CartRequest
	if err := c.ShouldBindJSON(&cartRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	cart, err := h.cartClient.RemoveFromCart(context.Background(), &cartRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *Handlers) UpdateCartHandler(c *gin.Context) {
	var cartRequest proto.CartRequest
	if err := c.ShouldBindJSON(&cartRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := h.cartClient.UpdateCart(context.Background(), &cartRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cart)
}
