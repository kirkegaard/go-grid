package server

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

// Redis says i need context so i gave it context
var ctx = context.Background()

// Redis client
var rdb *redis.Client

func InitRedis() {
	// Get Redis connection details from environment variables
	redisHost := GetEnv("REDIS_HOST", "localhost")
	redisPort := GetEnv("REDIS_PORT", "6379")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}
