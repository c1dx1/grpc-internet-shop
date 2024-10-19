package sessions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

func GenerateSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func CreateSession(ctx context.Context, redisClient *redis.Client, userID int64) (string, error) {
	sessionID, err := GenerateSessionID()
	if err != nil {
		log.Printf("session: error generate session id: %v", err)
		return "", err
	}

	err = redisClient.Set(ctx, sessionID, userID, time.Hour*720).Err()
	if err != nil {
		log.Printf("session: error set session id to redis: %v", err)
		return "", err
	}

	return sessionID, nil
}

func DeleteSession(ctx context.Context, redisClient *redis.Client, sessionID string) error {
	result := redisClient.Del(ctx, sessionID)
	if result.Err() != nil {
		log.Printf("session: deletesession: error delete session id from redis: %v", result.Err())
		return result.Err()
	}

	if result.Val() == 0 {
		log.Printf("session: deletesession: no session found")
	}

	return nil
}

func GetUserIDFromSession(ctx context.Context, redisClient *redis.Client, sessionID string) (string, error) {
	userID, err := redisClient.Get(ctx, sessionID).Result()
	if err == redis.Nil {
		return "", err
	} else if err != nil {
		return "", err
	}

	return userID, nil
}
