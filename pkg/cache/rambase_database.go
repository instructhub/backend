package cache

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
)

// InitializeRedis initializes the Redis client.
func init() {
	// Read Redis database number.
	dbNumber := 0

	host := os.Getenv("CACHE_HOST")
	port := os.Getenv("CACHE_PORT")
	if host == "" || port == "" {
		log.Fatalf("CACHE_HOST or CACHE_PORT is not set")
	}

	url := fmt.Sprintf("%s:%s", host, port)

	// Set Redis options.
	options := &redis.Options{
		Addr:         url,
		Password:     os.Getenv("CACHE_PASSWORD"),
		DB:           dbNumber,
		PoolSize:     10,
		MinIdleConns: 2,
	}

	RedisClient = redis.NewClient(options)

	// Test connection
	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v\n", err)
	}
}
