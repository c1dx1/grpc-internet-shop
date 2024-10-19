package main

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/metadata"
	"internet-shop/services/user-service/interceptors"
	"net/http"
	"strings"
)

func sessionMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			c.Abort()
			return
		}

		if strings.HasPrefix(token, "Bearer ") {
			sessionID := strings.TrimPrefix(token, "Bearer ")

			valid, err := interceptors.ValidateSession(c.Request.Context(), redisClient, sessionID)
			if err != nil || !valid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session or " + err.Error()})
				c.Abort()
				return
			}

			md := metadata.Pairs("session-id", sessionID, "authenticated", "true")
			newCtx := metadata.NewOutgoingContext(c.Request.Context(), md)
			c.Request = c.Request.WithContext(newCtx)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}
	}
}
